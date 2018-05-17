package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	md5Quota := make(chan struct{}, 1) //md5 можно держать только один
	wg := &sync.WaitGroup{}
	for val := range in {
		//Преобразуем входные данные в строки
		dataInt, ok := val.(int)
		if !ok {
			log.Println("SingleHash: cant convert result data to int")
			continue
		}
		data := strconv.Itoa(dataInt)

		wg.Add(1)
		go func() {
			defer wg.Done()
			crc32s := make(chan string)    // появится результат crc32(data)
			crc32md5s := make(chan string) // появится результат crc32(md5(data))

			go func() {
				crc32 := DataSignerCrc32(data)
				log.Printf("SingleHash %s: crc32 %s", data, crc32)
				crc32s <- crc32
			}()

			go func() {
				md5Quota <- struct{}{} // берем слот
				md5 := DataSignerMd5(data)
				log.Printf("SingleHash %s: md5 %s", data, md5)
				<-md5Quota // возвращаем слот

				crc32md5 := DataSignerCrc32(md5)
				log.Printf("SingleHash %s: crc32md5 %s", data, crc32md5)
				crc32md5s <- crc32md5
			}()

			// Собираем хеш, по мере готовности компонент
			var buff bytes.Buffer
			buff.WriteString(<-crc32s)
			buff.WriteByte('~')
			buff.WriteString(<-crc32md5s)

			result := buff.String()
			log.Printf("SingleHash %s: result %s", data, result)
			out <- result
		}()
	}

	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	const hashPartsNum = 6

	wg := &sync.WaitGroup{}
	for val := range in {
		data, ok := val.(string)
		if !ok {
			log.Println("MultiHash: cant convert result data to string")
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			inWG := &sync.WaitGroup{}
			parts := make([]string, hashPartsNum) //hashPartsNum компонент результата

			for th := 0; th < hashPartsNum; th++ {
				inWG.Add(1)
				go func(t int) {
					defer inWG.Done()
					crc32 := DataSignerCrc32(strconv.Itoa(t) + data)
					log.Printf("MultiHash %s: step %d: result %s", data, t, crc32)
					// Точно уверены, что к каждому эл-ту parts обратится только одна горутина.
					parts[t] = crc32
				}(th)
			}

			inWG.Wait()
			result := strings.Join(parts, "")
			log.Printf("MultiHash %s: result %s", data, result)

			out <- result
		}()
	}

	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var buff bytes.Buffer
	hashs := make([]string, 0)

	for val := range in {
		data, ok := val.(string)
		if !ok {
			log.Println("SingleHash: cant convert result data to string")
			continue
		}

		log.Printf("CombineResults: got %s", data)
		hashs = append(hashs, data)
	}

	sort.Strings(hashs)

	for _, str := range hashs {
		buff.WriteString(str)
		buff.WriteByte('_')
	}

	result := buff.String()
	out <- result[:len(result)-1]
}

func ExecutePipeline(jobs ...job) {
	var in chan interface{}
	out := make(chan interface{})

	var wg sync.WaitGroup
	for _, jb := range jobs {
		wg.Add(1)
		go func(j job, i, o chan interface{}) {
			defer wg.Done()
			j(i, o)
			close(o) //j закинула в канал все данные
		}(jb, in, out)

		in = out
		out = make(chan interface{})
	}

	wg.Wait()
}

func main() {
	logs, err := os.Create("./signer_log")
	if err != nil {
		log.Fatal(err)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(logs)

	inputData := make([]int, 1000)
	for i := range inputData {
		inputData[i] = i
	}
	// inputData := []int{0, 1, 1, 2, 3, 5, 8}
	//inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				log.Fatal("cant convert result data to string")
			}

			fmt.Printf("Result %s\n", data)
		}),
	}

	ExecutePipeline(hashSignJobs...)
}
