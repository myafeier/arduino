package hjscanner

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/myafeier/log"
	"github.com/pkg/errors"
)

var Instructions []Instruction

// 指令
type Instruction struct {
	Title       string      `json:"title,omitempty"`        //指令名
	Key         string      `json:"key,omitempty"`          //指令key
	ParamAmount int         `json:"param_amount,omitempty"` //参数数量
	Req         string      `json:"-"`                      //请求
	Resp        string      `json:"-"`                      //返回结果
	respChan    chan string `json:"-"`                      //命令执行结果
}

// 读取序列号
var InstructionOfReadMachineSn = Instruction{
	Title:       "读取序列号",
	Key:         "read_sn",
	ParamAmount: 0,
	Req:         "#*,read_machine_sn,0,0,0,0,0,*#",
	Resp:        "",
	respChan:    make(chan string),
}

// 读取硬件版本
var InstructionOfReadHwVer = Instruction{
	Title:       "读取硬件版本号",
	Key:         "read_hw_ver",
	ParamAmount: 0,
	Req:         "#*,read_read_harware_ver,0,0,0,0,0,*#",
	Resp:        "",
	respChan:    make(chan string),
}

// 读取固件版本号
var InstructionOfReadFwVer = Instruction{
	Title:       "读取固件版本号",
	Key:         "read_fw_ver",
	ParamAmount: 0,
	Req:         "#*,read_firmware_ver,0,0,0,0,0,*#",
	Resp:        "",
	respChan:    make(chan string),
}

// 测试通讯是否成功
var InstructionOfTestComminution = Instruction{
	Title:       "通讯测试",
	Key:         "test",
	ParamAmount: 0,
	Req:         "#*,test_comm,0,0,0,0,0,*#",
	Resp:        "test_comm_ok",
	respChan:    make(chan string),
}

// 仪器初始化命令
var InstructionOfInit = Instruction{
	Title:       "设备初始化",
	Key:         "init",
	ParamAmount: 0,
	Req:         "#*,machine_init,0,0,0,0,0,*#",
	Resp:        "machine_init_ok",
	respChan:    make(chan string),
}

// 调焦 参数1，移动距离，浮点型
var InstructionOfMoveZ = Instruction{
	Title:       "调焦",
	Key:         "zoom",
	ParamAmount: 1,
	Req:         "#*,move_s,%.2f,0,0,0,0,*#",
	Resp:        "move_s_ok",
	respChan:    make(chan string),
}

// 平移 参数1: x轴移动距离,浮点型,最大120;参数1: y轴移动距离,浮点型,最大80;
var InstructionOfMoveXY = Instruction{
	Title:       "平移",
	Key:         "move",
	ParamAmount: 2,
	Req:         "#*,move_xy,%.2f,%.2f,0,0,0,*#",
	Resp:        "move_xy_ok",
	respChan:    make(chan string),
}

// 进舱
var InstructionOfMoveIn = Instruction{
	Title:       "进仓",
	Key:         "diskin",
	ParamAmount: 0,
	Req:         "#*,move_in,0,0,0,0,0,*#",
	Resp:        "move_in_ok",
	respChan:    make(chan string),
}

// 出舱
var InstructionOfMoveOut = Instruction{
	Title:       "出仓",
	Key:         "diskout",
	ParamAmount: 0,
	Req:         "#*,move_out,0,0,0,0,0,*#",
	Resp:        "move_out_ok",
	respChan:    make(chan string),
}

// 打开激光,参数1： green / red
var InstructionOfOpenLaser = Instruction{
	Title:       "激光开",
	Key:         "openlaser",
	ParamAmount: 1,
	Req:         "#*,open_laser,%s,0,0,0,0,*#",
	Resp:        "open_laser_ok",
	respChan:    make(chan string),
}

// 关闭激光
var InstructionOfCloseLaser = Instruction{
	Title:       "激光关",
	Key:         "closelaser",
	ParamAmount: 1,
	Req:         "#*,close_laser,%s,0,0,0,0,*#",
	Resp:        "close_laser_ok",
	respChan:    make(chan string),
}

func init() {
	Instructions = append(Instructions, InstructionOfInit)
	Instructions = append(Instructions, InstructionOfTestComminution)
	Instructions = append(Instructions, InstructionOfMoveIn)
	Instructions = append(Instructions, InstructionOfMoveOut)
	Instructions = append(Instructions, InstructionOfMoveXY)
	Instructions = append(Instructions, InstructionOfMoveZ)
	Instructions = append(Instructions, InstructionOfOpenLaser)
	Instructions = append(Instructions, InstructionOfCloseLaser)
}

// 指令校验和编译
func (s Instruction) compile(param ...interface{}) (bytes []byte, err error) {
	if s.Req == InstructionOfMoveZ.Req {
		if len(param) != 1 {
			err = errors.WithStack(fmt.Errorf("invalid param amount:%+v", param))
			return
		}

		var t float64
		switch param[0].(type) {
		case string:
			t, err = strconv.ParseFloat(param[0].(string), 64)
			if err != nil {
				err = errors.WithStack(err)
				return
			}
		case float32:
			t = float64(param[0].(float32))
		case float64:
			t = param[0].(float64)
		default:
			err = errors.WithStack(fmt.Errorf("invalid param type"))
			return

		}
		if t > 1000 {
			err = errors.WithStack(fmt.Errorf("invalid param value"))
			return
		}
		param[0] = t

	} else if s.Req == InstructionOfMoveXY.Req {
		if len(param) != 2 {
			err = errors.WithStack(fmt.Errorf("invalid param amount"))
			return
		}

		var x, y float64
		switch param[0].(type) {
		case string:
			x, err = strconv.ParseFloat(param[0].(string), 64)
			if err != nil {
				err = errors.WithStack(err)
				return
			}
		case float32:
			x = float64(param[0].(float32))
		case float64:
			x = param[0].(float64)
		case int:
			x = float64(param[0].(int))
		default:
			err = fmt.Errorf("invalid param type")
			err = errors.WithStack(err)
			return

		}
		switch param[1].(type) {
		case string:
			y, err = strconv.ParseFloat(param[1].(string), 64)
			if err != nil {
				err = errors.WithStack(err)
				return
			}
		case float32:
			y = float64(param[1].(float32))
		case float64:
			y = param[1].(float64)
		case int:
			x = float64(param[1].(int))

		default:
			err = fmt.Errorf("invalid param type")
			err = errors.WithStack(err)
			return
		}

		if x > 120 {
			err = fmt.Errorf("x不允许大于120")
			err = errors.WithStack(err)
			return
		}
		if x < 0 {
			err = fmt.Errorf("x不允许小于0")
			err = errors.WithStack(err)
			return
		}
		if y < 0 {
			err = fmt.Errorf("y不允许小于0")
			err = errors.WithStack(err)
			return
		}
		param[0] = x
		param[1] = y
	} else if s.Req == InstructionOfOpenLaser.Req {
		if len(param) != 1 {
			err = fmt.Errorf("invalid param amount")
			err = errors.WithStack(err)
			return
		}
		if v, ok := param[0].(string); !ok {
			err = fmt.Errorf("invalid param 1 type")
			err = errors.WithStack(err)
			return
		} else {
			if v != "green" && v != "red" {
				err = fmt.Errorf("invalid param 1 value")
				err = errors.WithStack(err)
				return
			}
		}
	}

	bytes = []byte(fmt.Sprintf(s.Req, param...))
	return
}

var mutex sync.Mutex

// 指令执行必须保证处于单线程状态,
func (s Instruction) DoWithTimeout(conn io.ReadWriteCloser, ctx context.Context, params ...interface{}) (resp string, err error) {
	mutex.Lock()
	defer func() {
		mutex.Unlock()
		if err1 := recover(); err1 != nil {
			err = errors.WithStack(fmt.Errorf("%v", err1))
			return
		}
	}()

	//发送指令
	var ins []byte
	if ins, err = s.compile(params...); err != nil {
		return
	} else {
		if _, err = conn.Write(ins); err != nil {
			err = errors.WithStack(err)
			return
		} else {
			log.Debug("成功发送指令:%s", ins)
		}
	}
	for {
		select {
		case <-ctx.Done():
			err = fmt.Errorf("指令执行超时")
			err = errors.WithStack(err)
			return
		case resp = <-s.respChan:
			if strings.Contains(resp, s.Resp) {
				log.Debug("指令成功执行:%s", resp)
				return
			}
		}
	}
}
