package utils

import (
	"encoding/json"
	"net"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

// Struct2Bytes converts a struct to a byte slice using JSON encoding.
func Bytes2Struct[T any](data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func Struct2Bytes[T any](data T) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Bytes2ProtobufStruct converts protobuf bytes to a struct
func Bytes2ProtobufStruct[T proto.Message](data []byte) (T, error) {
	var result T
	// 创建一个新的实例
	resultType := reflect.TypeOf(result).Elem()
	newResult := reflect.New(resultType).Interface().(T)

	err := proto.Unmarshal(data, newResult)
	if err != nil {
		return result, err
	}
	return newResult, nil
}

// ProtobufStruct2Bytes converts a protobuf struct to bytes
func ProtobufStruct2Bytes[T proto.Message](data T) ([]byte, error) {
	bytes, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// 检查IP地址格式是否正确
func checkIp(ipStr string) bool {
	address := net.ParseIP(ipStr)
	if address == nil {
		// fmt.Println("ip地址格式不正确")
		return false
	} else {
		// fmt.Println("正确的ip地址", address.String())
		return true
	}
}

// IP地址转为Int64
func NetAton(ipStr string) int64 {
	bits := strings.Split(ipStr, ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func IsInnerIp(ipStr string) bool {
	if !checkIp(ipStr) {
		return false
	}
	inputIpNum := NetAton(ipStr)
	innerIpA := NetAton("10.255.255.255")
	innerIpB := NetAton("172.16.255.255")
	innerIpC := NetAton("192.168.255.255")
	innerIpD := NetAton("100.64.255.255")
	innerIpF := NetAton("127.255.255.255")

	return inputIpNum>>24 == innerIpA>>24 || inputIpNum>>20 == innerIpB>>20 ||
		inputIpNum>>16 == innerIpC>>16 || inputIpNum>>22 == innerIpD>>22 ||
		inputIpNum>>24 == innerIpF>>24
}
