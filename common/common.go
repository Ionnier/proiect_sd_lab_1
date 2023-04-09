package common

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ReturnInts struct {
	Numbers []int
}

type ReturnStrings struct {
	Words []string
}

type ReturnFloats struct {
	Numbers []float64
}

type test struct {
	url string
}

func (m test) Write(p []byte) (n int, err error) {
	a := fiber.AcquireAgent()
	data := make(map[string]string)
	data["message"] = string(p)
	a.JSON(data)
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(m.url)
	if err := a.Parse(); err != nil {
		panic(err)
	}
	if _, _, err := a.Bytes(); err != nil {
		fmt.Print(err)
	}
	return fmt.Print(string(p))
}

func GetHTTPLogger(url string) test {
	var asd test
	asd.url = url
	return asd
}
