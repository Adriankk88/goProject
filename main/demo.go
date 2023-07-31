package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
)

func main() {
	writeToFile(100)
}

type Emails struct {
	PostId int    `json:"postId"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

func (this *Emails) getEmail() string {
	return this.Email
}

func writeToFile(n int) {
	file, err := os.OpenFile("main/data.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open file err:", err)
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// 创建带缓存的写入器
	bufWriter := bufio.NewWriter(file)
	var wg sync.WaitGroup
	limitChan := make(chan struct{}, runtime.GOMAXPROCS(runtime.NumCPU())) // 最大并发协程数
	var mutex sync.Mutex
	for i := 1; i < n+1; i++ { // 写100行测试
		limitChan <- struct{}{}
		wg.Add(1)
		response := getEmailById(i)

		go func() {
			defer func() {
				if e := recover(); e != nil {
					fmt.Printf("WriteDataToTxt panic: %v,stack: %s\n", e, debug.Stack())
					// return
				}

				wg.Done()
				<-limitChan
			}()

			// 根据业务逻辑：先整合所有数据，然后再统一写WriteString()
			var emailStr string
			for j := 0; j < len(response); j++ {
				email := response[j].getEmail()
				emailStr += email + "\n"
			}

			// 要加锁/解锁，否则 bufWriter.WriteString 写入数据有问题
			mutex.Lock()
			_, err := bufWriter.WriteString(emailStr)
			//_, err := bufWriter.WriteString(strId + strName + strScore + "\n")
			if err != nil {
				fmt.Printf("WriteDataToTxt WriteString err: %v\n", err)
				return
			}
			mutex.Unlock()

		}()
		// bufWriter.Flush() // 刷入磁盘（错误示例：WriteDataToTxt err: short write，short write；因为循环太快，有时写入的数据量太小了）
		wg.Wait()
		err := bufWriter.Flush()
		if err != nil {
			return
		}
	}
}

func getEmailById(n int) []Emails {
	var url = "https://jsonplaceholder.typicode.com/posts/" + fmt.Sprintf("%d", n) + "/comments"
	resp, err := http.Get(url)
	var response []Emails
	if err != nil {
		fmt.Println("resp is error")
		return nil
	} else {
		content, _ := io.ReadAll(resp.Body)
		err := json.Unmarshal(content, &response)
		if err != nil {
			return nil
		}
		return response
	}
}
