package express

import (
	"fmt"
	"sync/atomic"
	"time"
)

func benchExpress(desc string, regex string, duration time.Duration) {
	exp := Compile(regex)
	count, ticker, stop := int32(0), time.NewTicker(duration), false

	fmt.Printf("Now Benching on %s: \n", desc)
	for i := 0; i < 10; i++ {
		fmt.Printf(" - %s \n", exp.Get())
	}

	<-ticker.C
	for !stop {
		select {
		case <-ticker.C:
			stop = true
		default:
			exp.Get()
			go atomic.AddInt32(&count, 1)
		}
	}

	var qps float64 = float64(count) / duration.Seconds()
	fmt.Printf("%s: Found qps to be %.1f for regex %s tested over %.1fs \n", desc, qps, regex, duration.Seconds())
}

func Bench() {
	const duration time.Duration = time.Second * 5

	//	benchExpress("emails", "[a-zA-Z0-9]+{30}@[a-zA-Z]+{10}\\.com(\\.au)?", duration)
	//	benchExpress("usernames", "[a-z]+{8}[0-9]*{2}", duration)
	//	benchExpress("ipv4", "[0-9]+{3}\\.[0-9]+{3}\\.[0-9]+{3}(:[0-9]+{4})?", duration)
	benchExpress("dates", "([1-9][0-9]*{3}-(1[0-2])|([1-9])-([1-9])|([1-2][0-9])|(30))#a #a #a", duration)

}
