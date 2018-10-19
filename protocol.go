package shuttle

import (
	"io"
	"bytes"
	"encoding/binary"
)

const (
	PacketHeadSize           = 20
	PacketMaxDataSize uint32 = 1024 * 1024
)

type PacketHead struct {
	Cmd     int32 //命令
	Act     int32 //动作
	Len     int32 //数据长度
	Index   int32 //序号
	Err     int32 //错误号
	Version int32 //版本
}

type Packet struct {
	Head *PacketHead //消息头，可能为nil
	Data string      //消息数据
}

type Protocol interface {
	ReadPacket(reader io.Reader, msg2 chan []byte) ([]byte, error)
	WritePacket(writer io.Writer, msg []byte) error
}

func (ep *Packet) ReadPacket(conn io.Reader, readerChannel chan []byte) ([]byte, error) {
	//声明一个临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)

	buffer := make([]byte, PacketMaxDataSize)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return nil, err
		}
		tmpBuffer = ep.UnPacket(append(tmpBuffer, buffer[:n]...), readerChannel)
	}
	return nil, nil
}

func (ep *Packet) WritePacket(writer io.Writer, msg []byte) error {
	_, err := writer.Write(msg)
	return err
}

//封包
func (p *Packet) Packet() []byte {
	np := make([]byte, 0)
	np = append(np, IntToBytes(p.Head.Cmd)...)
	np = append(np, IntToBytes(p.Head.Act)...)
	np = append(np, IntToBytes(p.Head.Len)...)
	np = append(np, IntToBytes(p.Head.Index)...)
	np = append(np, IntToBytes(p.Head.Err)...)
	np = append(np, []byte(p.Data)...)
	return np
}

//解包
func (p *Packet) UnPacket(buffer []byte, readerChannel chan []byte) []byte {
	for {
		if len(buffer) < PacketHeadSize {
			break
		}
		length := BytesToInt(buffer[8:12])
		readerChannel <- buffer[:PacketHeadSize+length]
		//fmt.Println(buffer[:20+length])
		buffer = buffer[PacketHeadSize+length:]
	}
	return buffer
}

func ReadPacket(b []byte) *Packet {
	cmd := BytesToInt(b[0:4])
	act := BytesToInt(b[4:8])
	length := BytesToInt(b[8:12])
	index := BytesToInt(b[12:16])
	err := BytesToInt(b[16:20])
	return NewPacket(cmd, act, index, err, string(b[20:20+length]))
}

func NewPacket(cmd, act int32, index, err int32, data string) *Packet {
	return &Packet{
		Head: &PacketHead{
			Cmd:   cmd,
			Act:   act,
			Len:   int32(len(data)),
			Index: index,
			Err:   err,
		},
		Data: data,
	}
}

//整形转换成字节
func IntToBytes(n int32) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.LittleEndian, &x)
	return int32(x)
}
