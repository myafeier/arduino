package hjscanner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/myafeier/log"
	"github.com/pkg/errors"
	"github.com/tarm/serial"
)

type ScannerStatus int32

const (
	ScannerStatusOfOk         ScannerStatus = 1
	ScannerStatusOfLost       ScannerStatus = -1
	ScannerStatusOfConnecting ScannerStatus = 2
)

var DefaultScaner *Scanner

func InitDefaultScanner() (sn string, err error) {
	dev := "/dev/hjscanner"
	DefaultScaner, err = InitScanner(dev)
	if err != nil {
		return
	}
	sn, err = DefaultScaner.RunInstruction(InstructionOfReadMachineSn)
	snSlice := bytes.Split([]byte(sn), []byte{':'})
	if len(snSlice) == 2 {
		sn = hex.EncodeToString(snSlice[1])
	} else {
		sn = "undefined"
	}
	return
}

func InitScanner(dev string) (scanner *Scanner, err error) {
	scanner = &Scanner{
		Port:        dev,
		reconn:      make(chan bool),
		reconnected: make(chan bool),
	}
	err = scanner.Connect()
	if err != nil {
		return
	}
	go scanner.Daemon()

	return
}

var atomValue atomic.Value

var ScannerStatusMap = map[ScannerStatus]string{
	ScannerStatusOfOk:         "状态正常",
	ScannerStatusOfLost:       "失去连接",
	ScannerStatusOfConnecting: "正在尝试连接",
}

func (s ScannerStatus) String() string {
	if str, ok := ScannerStatusMap[s]; ok {
		return str
	} else {
		return "-"
	}
}

const InstructionTimeout = 10 * time.Second //指令执行超时时间

type Scanner struct {
	Port        string             `json:"port,omitempty"` //设备USB端口
	Conn        io.ReadWriteCloser `json:"-"`
	Status      ScannerStatus      `json:"status,omitempty"`
	reconn      chan bool          `json:"-"` //开始重连
	reconnected chan bool          `json:"-"` //重联成功
	Watcher     []chan string      `json:"watcher,omitempty"`
}

// 尝试初始化设备并返回设备编码
func (s *Scanner) Init() (sn string, err error) {
	_, err = s.RunInstruction(InstructionOfInit)
	if err != nil {
		log.Error(err.Error())
		return
	}
	sn, err = DefaultScaner.RunInstruction(InstructionOfReadMachineSn)
	if err != nil {
		log.Error(err.Error())
		return
	}
	snSlice := bytes.Split([]byte(sn), []byte{':'})
	if len(snSlice) == 2 {
		sn = hex.EncodeToString(snSlice[1])
	} else {
		sn = "undefined"
	}
	return
}

func (s *Scanner) SetPort(port string) {
	s.Port = port
}

func (s *Scanner) GetState() ScannerStatus {
	return s.Status
}
func (s *Scanner) SetState(state ScannerStatus) {
	s.Status = state
}

func (s *Scanner) Connect() (err error) {
	log.Debug("trying connect ard...")
	cfg := new(serial.Config)
	cfg.Name = s.Port
	cfg.Baud = 115200
	s.Conn, err = serial.OpenPort(cfg)
	if err != nil {
		log.Debug("trying connect fail:%s", err.Error())
		s.SetState(ScannerStatusOfLost)
		return errors.WithStack(err)
	} else {
		log.Debug("arduino connected")
		s.SetState(ScannerStatusOfOk)
		return nil
	}
}

// 后台监控进程
func (s *Scanner) Daemon() {
	go s.Read()
	reconnecting := false
	for {
		select {
		case <-s.reconn:
			log.Debug("收到重连信号")
			if reconnecting {
				continue
			} else {
				reconnecting = true
			}
			log.Debug("重试连接arduino")
			//重试连接
			if err := s.Connect(); err != nil {
				log.Error("arduitrueno重试连接失败:%s", err.Error())
				go func() {
					time.Sleep(1 * time.Second)
					s.reconn <- true
				}()
			} else {
				s.reconnected <- true
				log.Debug("已发送连接成功信号")
			}
			reconnecting = false
		}
	}
}

// 运行指令
//
//	一个指令发送后，会通过daemon监控运行结果，或超时返回error
func (s *Scanner) RunInstruction(instruction Instruction, params ...interface{}) (resp string, err error) {
	state := s.GetState()
	if state != ScannerStatusOfOk && InstructionOfInit.Req != instruction.Req {
		s.reconn <- true
		s.SetState(ScannerStatusOfLost)
		err = errors.WithStack(fmt.Errorf("ard设备状态异常:(%s),请检查连接或稍后重试", s.Status.String()))
		return
	}
	s.Watcher = append(s.Watcher, instruction.respChan)
	ctx, cancelFunc := context.WithTimeout(context.Background(), InstructionTimeout)
	defer func() {
		cancelFunc() //如果
		//从观察者中去掉当前指令
		for k, v := range s.Watcher {
			if v == instruction.respChan {
				s.Watcher = append(s.Watcher[:k], s.Watcher[k+1:]...)
			}
		}

	}()
	if resp, err = instruction.DoWithTimeout(s.Conn, ctx, params...); err != nil {
		if strings.Contains(err.Error(), "input/output error") {
			s.SetState(ScannerStatusOfLost)
			s.reconn <- true
			err = fmt.Errorf("扫描仪设备输入输出错误，请检查电缆是否连接正确")
		}
		log.Error(err.Error())
		return
	}
	return
}

var readMutex sync.Mutex

// 从连接读取消息，发送给观察者
func (s *Scanner) Read() {
	readMutex.Lock()
	defer func() {
		readMutex.Unlock()
	}()
	reader := bufio.NewReader(s.Conn)

	for {
		select {
		case <-s.reconnected:
			log.Debug("reset readbuffer")
			reader = bufio.NewReader(s.Conn)
		default:
			time.Sleep(1 * time.Millisecond)
			if bytes, _, err := reader.ReadLine(); err != nil {
				if err != io.EOF {
					log.Error("读取Arduino失败： %s", err.Error())
				}
			} else {
				log.Debug("read:  %s \n", bytes)
				for _, w := range s.Watcher {
					w <- string(bytes)
				}
			}
		}
	}

}

func RunInstruction(cmd string, params []interface{}) (err error) {
	switch cmd {
	case "test":
		_, err = DefaultScaner.RunInstruction(InstructionOfTestComminution, params...)
	case "init":
		_, err = DefaultScaner.RunInstruction(InstructionOfInit, params...)
	case "move":
		_, err = DefaultScaner.RunInstruction(InstructionOfMoveXY, params...)
	case "zoom":
		_, err = DefaultScaner.RunInstruction(InstructionOfMoveZ, params...)
	case "diskin":
		_, err = DefaultScaner.RunInstruction(InstructionOfMoveIn, params...)
	case "diskout":
		_, err = DefaultScaner.RunInstruction(InstructionOfMoveOut, params...)
	case "openlaser":
		_, err = DefaultScaner.RunInstruction(InstructionOfOpenLaser, params...)
	case "closelaser":
		_, err = DefaultScaner.RunInstruction(InstructionOfCloseLaser, params...)

	default:
		err = fmt.Errorf("unsupported instruction")
	}
	return
}
