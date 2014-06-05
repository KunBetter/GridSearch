// GenTestData
package GridSearch

import (
	"math/rand"
	"time"
)

func GenRandomLo() int32 {
	rand_ := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand_.Int31n(CHINA_RECT.Right-CHINA_RECT.Left) + CHINA_RECT.Left
}

func GenRandomLa() int32 {
	rand_ := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand_.Int31n(CHINA_RECT.Top-CHINA_RECT.Bottom) + CHINA_RECT.Bottom
}

func GenRandomID() int32 {
	rand_ := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand_.Int31n(10000) * rand_.Int31n(20000) % 1234567890
}

func GenRandomRect() *rect {
	rand1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	h := GRID_TOP_HEIGHT / (rand1.Int31n(100) + 10)
	rand2 := rand.New(rand.NewSource(time.Now().UnixNano()))
	w := GRID_TOP_WIDTH / (rand2.Int31n(100) + 10)

	lo := GenRandomLo()
	la := GenRandomLa()
	return &rect{
		lo,
		la,
		lo + w,
		la - h,
	}
}
