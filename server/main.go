package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/cmplx"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"example.com/common"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/emirpasic/gods/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/streadway/amqp"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

func emptyCharactersToEx6(inputs []string) []string {
	returnValue := make([]string, 0)
	for _, val := range inputs {
		el := strings.TrimSpace(val)
		if el == "" {
			continue
		}
		returnValue = append(returnValue, el)
	}
	return returnValue
}

type config struct {
	messaging_server string
	server_port      string
	queue_name       string
	rpc_port         string
	logstash         string
}

func RSA_OAEP_Encrypt(secretMessage string, key rsa.PublicKey) string {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, &key, []byte(secretMessage), label)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

var clients map[string]string = make(map[string]string)

type NetRPCListener string

func (nrl *NetRPCListener) Exercitiul2(array []string, reply *int) error {
	var wg sync.WaitGroup
	isPrimeChan := make(chan bool)
	for _, val := range array {
		wg.Add(1)
		go func(word string, writeChan chan bool) {
			defer wg.Done()
			current := ""
			for _, val2 := range word {
				switch string(val2) {
				case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
					current += string(val2)
				}
			}
			val, err := strconv.Atoi(current)
			if err != nil {
				return
			}
			if isPP(val) {
				writeChan <- true
			}
		}(val, isPrimeChan)
	}

	count := make(chan int)

	go func(writeCount chan int, readNumbers chan bool) {
		count := 0
		for number := range readNumbers {
			if number {
				count++
			}
		}
		writeCount <- count
	}(count, isPrimeChan)
	wg.Wait()
	close(isPrimeChan)
	*reply = <-count
	return nil
}

func (nrl *NetRPCListener) Exercitiul12(array []int, reply *int) error {
	var wg sync.WaitGroup
	writeChannel := make(chan int)
	for _, val := range array {
		wg.Add(1)
		go func(word int, writeChan chan int) {
			defer wg.Done()
			current := fmt.Sprintf("%d", word)
			current = string(current[0]) + current
			val, err := strconv.Atoi(current)
			if err != nil {
				return
			}
			writeChan <- val
		}(val, writeChannel)
	}

	count := make(chan int)

	go func(writeCount chan int, readNumbers chan int) {
		sum := 0
		for number := range readNumbers {
			sum += number
		}
		writeCount <- sum
	}(count, writeChannel)
	wg.Wait()
	close(writeChannel)
	*reply = <-count
	return nil
}

func ceaserCypher(orig string, left bool, number int) string {
	returnValue := make([]rune, 0)
	rotater := number
	if left {
		rotater *= (-1)
	}
	for _, val := range orig {
		xd := (int(val) - int('a') - 3) % 26
		if xd < 0 {
			xd += (int('z' + 1))
		} else {
			xd += int('a')
		}
		returnValue = append(returnValue, rune(xd))
	}
	return string(returnValue)
}

type list struct {
	data []string
}

func (nrl *NetRPCListener) Exercitiul6(array []string, reply *list) {
	var wg sync.WaitGroup
	finishedWordChannel := make(chan string)
	for i, v := range array {
		wg.Add(1)
		go func(writeChannel chan string, word string, position int) {
			defer wg.Done()
			list := strings.Split(word, " ")
			if len(list) > 3 {
				return
			}
			toCrypt := list[0]
			way := list[1]
			rotation, _ := strconv.Atoi(list[2])
			left := false
			if way == "STANGA" {
				left = true
			}
			resultString := ceaserCypher(toCrypt, left, rotation)
			writeChannel <- fmt.Sprintf("%d %s", position, resultString)
		}(finishedWordChannel, v, i)
	}

	returnMap := make(chan []string)
	go func(finishedWord chan string, returnMap chan []string) {
		returnValue := make([]string, 0)
		for value := range finishedWord {
			returnValue = append(returnValue, value)
		}
		returnMap <- returnValue

	}(finishedWordChannel, returnMap)
	wg.Wait()
	close(finishedWordChannel)

	returnValue := <-returnMap

	mapResult := getMapOfResults(returnValue)
	fmt.Println(reply, reply.data)
	reply.data = processMap(mapResult)
}

func (nrl *NetRPCListener) Exercitiul11(array []int, reply *int) error {
	var wg sync.WaitGroup
	k := array[0]
	numbers := array[1:]
	writeChannel := make(chan int)
	for _, val := range numbers {
		wg.Add(1)
		go func(word int, writeChan chan int) {
			defer wg.Done()
			wo := fmt.Sprint(word)
			for i := 0; i < k; i++ {
				wo = permutare1Dreapta(wo)
			}
			val, err := strconv.Atoi(wo)
			if err != nil {
				return
			}
			writeChan <- val
		}(val, writeChannel)
	}

	count := make(chan int)

	go func(writeCount chan int, readNumbers chan int) {
		sum := 0
		for number := range readNumbers {
			sum += number
		}
		writeCount <- sum
	}(count, writeChannel)
	wg.Wait()
	close(writeChannel)
	*reply = <-count
	return nil
}

func (nrl *NetRPCListener) Exercitiul13(array []int, reply *common.ReturnFloats) error {
	if len(array) <= 3 || len(array)%2 == 1 {
		return errors.New("vectorul nu este bun")
	}
	a := array[0]
	b := array[1]
	numbers := array[2:]
	var wg sync.WaitGroup
	writeGoodModules := make(chan float64)
	for i := 0; i < len(numbers); i++ {
		wg.Add(1)
		real := numbers[i]
		i++
		imaginar := numbers[i]
		go func(real int, imaginar int, writeChan chan float64) {
			defer wg.Done()
			c := complex(float64(real), float64(imaginar))
			val := cmplx.Abs(c)
			if float64(a) <= val && val <= float64(b) {
				return
			}
			writeChan <- val
		}(real, imaginar, writeGoodModules)
	}
	returnChan := make(chan []float64)
	go func(readChan chan float64, returnChan chan []float64) {
		hello := treeset.NewWith(utils.Float64Comparator)
		for number := range readChan {
			hello.Add(number)
		}
		asd := make([]float64, 0)
		for a := hello.Iterator(); a.Next(); {
			valueOfI, ok := a.Value().(float64)
			if ok {
				asd = append(asd, valueOfI)
			} else {
				asd = append(asd, -1)
			}
		}
		returnChan <- asd
	}(writeGoodModules, returnChan)
	wg.Wait()
	close(writeGoodModules)
	val := <-returnChan
	reply.Numbers = val
	return nil
}

type Comparator2 func(a, b float64) int

func permutare1Dreapta(value string) (returnValue string) {
	returnValue = string(value[len(value)-1]) + value[0:(len(value)-1)]
	return
}

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()
	configServer := config{}

	var file *os.File
	var err error
	file, err = os.Open("config.txt")
	if err != nil {
		file, err = os.Open("../config.txt")
		if err != nil {
			sysvar := strings.ToLower(os.Getenv("NON_INTERACTIVE"))
			switch sysvar {
			case "true", "t", "yes", "1", "y":
				panic(err)
			}
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				path := scanner.Text()
				file, err = os.Open(path)
				if err == nil {
					break
				}
			}
		}
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value := scanner.Text()
		words := strings.SplitN(value, "=", 2)
		switch words[0] {
		case "MESSAGING_SERVER":
			configServer.messaging_server = words[1]
		case "SERVER_PORT":
			configServer.server_port = words[1]
		case "QUEUE_NAME":
			configServer.queue_name = words[1]
		case "RPC_PORT":
			configServer.rpc_port = words[1]
		case "LOGSTASH":
			configServer.logstash = words[1]
		}
	}
	file.Close()
	log.SetOutput(common.GetHTTPLogger(configServer.logstash))

	go func() {
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%s", configServer.rpc_port))
		if err != nil {
			panic(err)
		}
		inbound, err := net.ListenTCP("tcp", address)
		if err != nil {
			panic(err)
		}
		listener := new(NetRPCListener)
		rpc.Register(listener)
		rpc.Accept(inbound)
	}()

	connection, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s/", configServer.messaging_server))
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	ch, err := connection.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	defer ch.Close()

	q, err := ch.QueueDeclare(
		configServer.queue_name, // name
		false,                   // durable
		false,                   // delete when unused
		false,                   // exclusive
		false,                   // no-wait
		nil,                     // arguments
	)

	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected to RabbitMQ instance")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			words := strings.SplitN(string(d.Body), " ", 3)
			if len(words) != 3 {
				continue
			}
			val := new(big.Int)
			val.SetString(words[0], 10)
			val2, e2 := strconv.Atoi(words[1])
			if e2 != nil {
				continue
			}
			publicKey := rsa.PublicKey{
				N: val,
				E: val2,
			}
			client_token := uuid.New().String()
			message := RSA_OAEP_Encrypt(client_token, publicKey)
			clients[client_token] = words[2]
			err = ch.Publish(
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(message),
				})
			if err != nil {
				log.Print(err)
			} else {
				log.Printf("Send authentification to %s", words[2])
			}

		}
	}()

	log.Println(configServer)

	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	app.Use(logger.New(logger.Config{
		Output: common.GetHTTPLogger(configServer.logstash),
	}))

	app.Use(func(c *fiber.Ctx) error {
		val, exists := c.GetReqHeaders()["Authorization"]

		if !exists {
			return fiber.NewError(fiber.StatusUnauthorized, "No authorization header.")
		}
		val, ok := clients[val]
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "Bad authentification.")
		}

		log.Printf("Client %s validat.\n", val)

		c.Locals("user", val)

		return c.Next()
	})

	// app.Use(func(c *fiber.Ctx) error {
	// 	auth := c.GetReqHeaders()
	// 	if val, ok := auth["Authorization"]; !ok {
	// 		return fiber.NewError(fiber.StatusUnauthorized, "No Authorizaion bearer name")
	// 	} else {
	// 		words := strings.Split(val, " ")
	// 		if words[0] != "Bearer" || len(words) != 2 {
	// 			return fiber.NewError(fiber.StatusUnauthorized, "No Bearer name")
	// 		}
	// 		log.Printf("Client %s conectat.\n", words[1])
	// 		c.Locals("user", words[1])
	// 	}
	// 	return c.Next()
	// })

	app.Post("/exercitiul1/", func(c *fiber.Ctx) error {
		var b map[string]string

		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}

		user := c.Locals("user")

		log.Printf("Clientul %s a facut request catre /exercitiul1 cu datele %s.\n", user, b["words"])

		words := strings.Split(b["words"], ",")
		var processedWords []string

		for _, v := range words {
			processedWords = append(processedWords, strings.Trim(v, " "))
		}

		writeChannel := make(chan string, len(processedWords))

		for i := 0; i < len(processedWords[0]); i++ {
			go extractLetterFrom(writeChannel, processedWords, i)
		}

		var returnValues []string

		for len(returnValues) != len(processedWords[0]) {
			log.Print(returnValues)

			returnValues = append(returnValues, <-writeChannel)
		}

		log.Print(returnValues)
		resultsAsMap := getMapOfResults(returnValues)
		log.Print(resultsAsMap)

		wordsInGoodOrder := processMap(resultsAsMap)
		log.Print(wordsInGoodOrder)

		data := strings.Join(wordsInGoodOrder, ", ")
		log.Printf("Server a procesat si trimite datele %s\n", data)
		return c.Status(fiber.StatusOK).JSON(newResponse(data))
	})

	app.Post("/exercitiul3/", func(c *fiber.Ctx) error {
		var b map[string]string

		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}

		log.Printf("Clientul %s a facut cerere cu datele %s", c.Locals("user"), b["numbers"])

		numbers := strings.Split(b["numbers"], ", ")

		if len(numbers) == 0 || len(strings.Trim(b["numbers"], " ")) == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "No numbers parameter in body")
		}

		reverseStrings := make(chan string, len(numbers))
		sumNumbers := make(chan int, len(numbers))

		for i, v := range numbers {
			go reverseNumber(v, i, reverseStrings, sumNumbers)
		}

		returnValueString := make(chan string)
		returnValueSum := make(chan int)

		go func(collectChannel chan string, sendChannel chan string, length int) {
			returnString := make([]string, 0, length)
			for len(returnString) != length {
				returnString = append(returnString, <-collectChannel)
			}
			resultMap := getMapOfResults(returnString)
			sendChannel <- strings.Join(processMap(resultMap), ", ")

		}(reverseStrings, returnValueString, len(numbers))

		go func(collectChannel chan int, sendChannel chan int, length int) {
			sum := 0
			for i := 0; i < length; i++ {
				sum += <-collectChannel
			}
			sendChannel <- sum

		}(sumNumbers, returnValueSum, len(numbers))

		message := fmt.Sprintf("%s cu suma %d", <-returnValueString, <-returnValueSum)
		log.Printf("Serverul a procesat datele si trimite %s", message)
		return c.Status(fiber.StatusOK).JSON(newResponse(message))
	})

	app.Post("/exercitiul4/", func(c *fiber.Ctx) error {
		var body map[string][]int

		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return err
		}

		log.Printf("Clientul %s a facut cerere cu datele %v", c.Locals("user"), body["numbers"])

		numbers := body["numbers"]
		a := numbers[0]
		b := numbers[1]
		sir := numbers[2:]

		validNumber := make(chan int, len(sir)/2+1)
		sum := make(chan float64)

		var wg sync.WaitGroup
		for _, v := range sir {
			wg.Add(1)
			go func(valid chan int, number int) {
				defer wg.Done()
				copy := number
				sum := 0
				for ; copy != 0; copy /= 10 {
					sum += copy % 10
				}
				if a <= sum && sum <= b {
					valid <- number
				}

			}(validNumber, v)
		}
		go func(valid chan int, sum chan float64) {
			l := 0
			count := 0
			for i := range valid {
				l += i
				count += 1
			}
			sum <- (float64(l) / float64(count))
		}(validNumber, sum)

		wg.Wait()
		close(validNumber)

		toSend := <-sum

		log.Printf("Server-ul a calculat %f", toSend)
		return c.Status(fiber.StatusOK).JSON(newResponse(fmt.Sprintf("%f", toSend)))
	})

	app.Post("/exercitiul5", func(c *fiber.Ctx) error {
		var b map[string][]string

		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}

		log.Printf("Clientul %s a facut cerere cu datele %v", c.Locals("user"), b["numbers"])

		var wg sync.WaitGroup

		valid := make(chan string, len(b["numbers"])/2)

		for i, v := range b["numbers"] {
			wg.Add(1)
			go func(writeChannel chan string, value string, index int) {
				defer wg.Done()
				if matched, err := regexp.Match("^[01]+$", []byte(value)); err == nil {
					if !matched {
						return
					}
					toHex := 0
					for i, v := range value {
						if v != '1' {
							continue
						}
						toHex += int(math.Pow(2, float64(len(value)-1-i)))
					}
					writeChannel <- fmt.Sprintf("%d %d", index, toHex)
				}
			}(valid, v, i)
		}
		resultString := make(chan string)

		go func(values chan string, returnValue chan string) {
			returnVal := make(map[int]int, 0)
			for val := range values {
				valI := strings.Split(val, " ")
				index, _ := strconv.Atoi(valI[0])
				val, _ := strconv.Atoi(valI[1])
				returnVal[index] = val
			}
			str := make([]string, 0)
			for _, val := range returnVal {
				str = append(str, fmt.Sprintf("%d", val))
			}
			returnValue <- strings.Join(str, " ")
		}(valid, resultString)
		wg.Wait()
		close(valid)
		data := <- resultString
		log.Printf("Serverul trimite catre %s datele %v", c.Locals("user"), data)
		return c.Status(fiber.StatusOK).JSON(newResponse(data))
	})

	app.Post("/exercitiul6", func(c *fiber.Ctx) error {
		var b map[string][]string

		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}
		log.Printf("Clientul %s a facut cerere cu datele %v", c.Locals("user"), b["words"])

		var wg sync.WaitGroup
		finishedWordChannel := make(chan string)
		for i, v := range b["words"] {
			wg.Add(1)
			go func(writeChannel chan string, word string, position int) {
				defer wg.Done()
				list := strings.Split(word, " ")
				if len(list) > 3 {
					return
				}
				toCrypt := list[0]
				way := list[1]
				rotation, _ := strconv.Atoi(list[2])
				left := false
				if way == "STANGA" {
					left = true
				}
				resultString := ceaserCypher(toCrypt, left, rotation)
				writeChannel <- fmt.Sprintf("%d %s", position, resultString)
			}(finishedWordChannel, v, i)
		}

		returnMap := make(chan []string)
		go func(finishedWord chan string, returnMap chan []string) {
			returnValue := make([]string, 0)
			for value := range finishedWord {
				returnValue = append(returnValue, value)
			}
			returnMap <- returnValue

		}(finishedWordChannel, returnMap)
		wg.Wait()
		close(finishedWordChannel)

		returnValue := <-returnMap

		mapResult := getMapOfResults(returnValue)
		var ret ReturnWords
		ret.Words = processMap(mapResult)
		log.Printf("Serverul trimite catre %s datele %v", c.Locals("user"), ret)
		return c.Status(fiber.StatusOK).JSON(ret)

	})
	app.Post("/exercitiul7", func(c *fiber.Ctx) error {
		var b map[string]string

		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}
		log.Printf("Clientul %s a facut cerere cu datele %v", c.Locals("user"), b["word"])


		receivedWord := b["word"]

		numbers := emptyCharactersToEx6(regexp.MustCompile(`[a-zA-Z]+`).Split(receivedWord, -1))
		letters := emptyCharactersToEx6(regexp.MustCompile(`[0-9]+`).Split(receivedWord, -1))

		writeChannel := make(chan string, len(numbers)/2)
		var wg sync.WaitGroup

		for i, v := range numbers {
			wg.Add(1)
			go func(number string, letter string, index int, writeChannel chan string) {
				defer wg.Done()
				i, _ := strconv.Atoi(number)
				str := ""
				for s := 0; s < i; s++ {
					str += letter
				}
				writeChannel <- fmt.Sprintf("%d %s", index, str)
			}(v, letters[i], i, writeChannel)
		}
		finalString := make(chan string)
		go func(writeChannel chan string, returnChannel chan string) {
			asd := make(map[int]string, 0)
			for val := range writeChannel {
				words := strings.SplitN(val, " ", 2)
				i, _ := strconv.Atoi(words[0])
				asd[i] = words[1]
			}
			words2 := sortMap(asd)
			returnChannel <- strings.Join(words2, "")
		}(writeChannel, finalString)
		wg.Wait()
		close(writeChannel)
		data := <- finalString
		log.Printf("Serverul trimite catre %s datele %v", c.Locals("user"), data)
		return c.Status(fiber.StatusOK).JSON(newResponse(data))
	})

	app.Post("/exercitiul8/", func(c *fiber.Ctx) error {
		var b map[string][]int

		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}

		receivedArray := b["numbers"]

		log.Printf("Server a primit %v", receivedArray)

		var wg sync.WaitGroup

		primeNumbers := make(chan int)

		for _, v := range receivedArray {
			wg.Add(1)
			go func(number int, writeChan chan int) {
				defer wg.Done()
				if isPrime(number) {
					writeChan <- len(fmt.Sprintf("%d", number))
				}
			}(v, primeNumbers)
		}
		sum := make(chan int)
		go func(writeChan chan int, returnChan chan int) {
			sum := 0
			for el := range writeChan {
				sum += el
			}
			returnChan <- sum
		}(primeNumbers, sum)

		wg.Wait()
		close(primeNumbers)
		message := fmt.Sprintf("%d", <-sum)
		log.Printf("Server-ul trimite mesajul %s", message)
		return c.Status(fiber.StatusOK).JSON(newResponse(message))

	})

	app.Post("/exercitiul9", func(c *fiber.Ctx) error {
		var b map[string][]string
		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}
		log.Printf("Server-ul a primit %v", b["words"])
		numberAddChan := make(chan bool)
		var wg sync.WaitGroup
		for _, val := range b["words"] {
			wg.Add(1)
			go func(isValid chan bool, word string) {
				defer wg.Done()
				for index, val := range word {
					if index%2 == 0 {
						switch strings.ToLower(string(val)) {
						case "a", "e", "i", "o", "u", "î", "â", "ă":
							// do nothing
						default:
							return
						}

					}
				}
				isValid <- true
			}(numberAddChan, val)
		}
		count := make(chan int)
		go func(validChannel chan bool, returnChannel chan int) {
			count := 0
			for r := range validChannel {
				if r {
					count += 1
				}
			}
			returnChannel <- count
		}(numberAddChan, count)

		wg.Wait()
		close(numberAddChan)
		mess := fmt.Sprintf("%d", <-count)
		log.Printf("Server-ul trimite %s", mess)
		return c.Status(fiber.StatusOK).JSON(newResponse(mess))

	})

	app.Post("/exercitiul10", func(c *fiber.Ctx) error {
		var b map[string][]string
		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}
		log.Printf("Server-ul a primit %v de la %s", b["numbers"], c.Locals("user"))
		var wg sync.WaitGroup
		validDivisors := make(chan int)
		for _, val := range b["numbers"] {
			if number, err := strconv.Atoi(val); err == nil {
				wg.Add(1)
				go func(nr int, valid chan int){
					defer wg.Done()
					for i:=1; i<int(math.Sqrt(float64(nr))); i++ {
						if (nr % i) == 0 {
							valid <- i;
							if (nr / i) != i {
								valid <- nr/i
							}
						}
					}
				}(number, validDivisors)
			}
		}
		cmmdc := make(chan int)
		go func(readChan chan int, writeChan chan int){
			divisors := make(map [int] int)
			for number := range readChan {
				val, ok := divisors[number];
				if !ok {
					divisors[val] = 0
				}
				divisors[number] = divisors[number] + 1
			}
			log.Printf("Server-ul a calculat toti divizorii %v de la %s", divisors, c.Locals("user"))
			maxKey := -1
			for key, val := range(divisors) {
				if maxKey < key && val >= divisors[1] {
					maxKey = key
				}
			}
			cmmdc <- maxKey
		}(validDivisors ,cmmdc)
		wg.Wait()
		close(validDivisors)
		mess := fmt.Sprintf("%d", <- cmmdc)
		log.Printf("Server-ul trimite %s", mess)
		return c.Status(fiber.StatusOK).JSON(newResponse(mess))
	})

	app.Post("/exercitiul14", func(c *fiber.Ctx) error {
		var b map[string][]string

		if err := json.Unmarshal(c.Body(), &b); err != nil {
			return err
		}

		var wg sync.WaitGroup
		finishedWordChannel := make(chan string)
		for i, v := range b["words"] {
			wg.Add(1)
			go func(writeChannel chan string, word string, position int) {
				defer wg.Done()
				err := passwordvalidator.Validate(word, 47)
				if err != nil {
					return
				}
				writeChannel <- fmt.Sprintf("%d %s", position, word)

			}(finishedWordChannel, v, i)
		}

		returnMap := make(chan []string)
		go func(finishedWord chan string, returnMap chan []string) {
			returnValue := make([]string, 0)
			for value := range finishedWord {
				returnValue = append(returnValue, value)
			}
			returnMap <- returnValue

		}(finishedWordChannel, returnMap)
		wg.Wait()
		close(finishedWordChannel)

		returnValue := <-returnMap

		mapResult := getMapOfResults(returnValue)
		var ret ReturnWords
		ret.Words = processMap(mapResult)
		return c.Status(fiber.StatusOK).JSON(ret)

	})
	listenRange := fmt.Sprintf(":%s", configServer.server_port)
	log.Printf("Server listening %s...", listenRange)
	app.Listen(listenRange)
	<-forever

}

type ReturnWords struct {
	Words []string `json:"words"`
}

func isPP(number int) bool {
	c := math.Sqrt(float64(number))
	if math.Mod(c, 1.0) == 0 {
		return true
	}
	return false
}

func isPrime(number int) bool {
	if number == 0 || number == 1 {
		return false
	}
	for i := 2; i <= number/2; i++ {
		if number%i == 0 {
			return false
		}
	}
	return true
}

func sortMap(og map[int]string) []string {
	task := make([]string, 0)
	keys := make([]int, 0)
	for k := range og {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, val := range keys {
		task = append(task, og[val])
	}
	return task
}

// func sortMapInts(og map[int]int) []int {
// 	task := make([]int, 0)
// 	keys := make([]int, 0)
// 	for k := range og {
// 		keys = append(keys, k)
// 	}
// 	sort.Ints(keys)
// 	for _, val := range keys {
// 		task = append(task, og[val])
// 	}
// 	return task
// }

func getMapOfResults(results []string) (returnValue map[int]string) {
	returnValue = make(map[int]string, len(results))
	for _, word := range results {
		word = strings.Trim(word, " ")
		mapValue := strings.SplitN(word, " ", 2)
		i, _ := strconv.Atoi(mapValue[0])
		returnValue[i] = mapValue[1]
	}
	return
}

func processMap(origMap map[int]string) (returnValue []string) {
	returnValue = make([]string, 0, len(origMap))
	for i := 0; i < len(origMap); i++ {
		returnValue = append(returnValue, origMap[i])
	}
	return
}

func extractLetterFrom(message chan string, words []string, i int) {
	returnValue := fmt.Sprintf("%d ", i)
	for _, v := range words {
		returnValue += string(v[i])
	}
	message <- returnValue
}

func reverseNumber(number string, index int, returnValue chan string, sendToSum chan int) {
	reverseNum := reverse(number)
	returnValue <- fmt.Sprintf("%d %s", index, reverseNum)
	i, _ := strconv.Atoi(reverseNum)
	sendToSum <- i
}

func reverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return
}

func customErrorHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	switch t := err.(type) {
	case *fiber.Error:
		code = t.Code
		message = t.Message
	case error:
		message = fmt.Sprint(t)
	}

	err = ctx.Status(code).JSON(newResponse(message))

	if err != nil {
		log.Println(err)
	}

	return nil
}

type response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func newResponse(message string, b ...interface{}) response {
	if len(b) == 0 {
		return response{
			Message: message}
	}
	return response{
		Message: message,
		Data:    b[0]}
}
