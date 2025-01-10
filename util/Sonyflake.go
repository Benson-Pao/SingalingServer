package util

import (
	"log"
	"time"

	"github.com/sony/sonyflake"
)

// 用來初始化雪花 ID 生成器
var sf *sonyflake.Sonyflake

func init() {
	// 設置雪花 ID 生成器的起始時間
	startTime := time.Date(2024, time.December, 25, 0, 0, 0, 0, time.UTC)
	var st sonyflake.Settings
	st.StartTime = startTime
	// 自定義機器 ID 不使用則預設mac address
	//st.MachineID = func() (uint16, error) {
	// 可以手動指定一個機器 ID，例如 1
	// 範圍是 0 ~ 65535
	//  return 1, nil
	//}
	sf = sonyflake.NewSonyflake(st)

	if sf == nil {
		log.Fatal("無法初始化雪花 ID 生成器")
	}
}

// 生成雪花 ID
func GenerateID() (uint64, error) {
	id, err := sf.NextID()
	if err != nil {
		return 0, err
	}
	return id, nil
}
