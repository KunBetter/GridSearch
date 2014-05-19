// GenTestData
package main

import (
	"math/rand"
	"time"
)

func genRandomLo() int32 {
	rand_ := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand_.Int31n(CHINA_RECT.Right-CHINA_RECT.Left) + CHINA_RECT.Left
}

func genRandomLa() int32 {
	rand_ := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand_.Int31n(CHINA_RECT.Top-CHINA_RECT.Bottom) + CHINA_RECT.Bottom
}

func genRandomID() int32 {
	rand_ := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rand_.Int31n(10000) * rand_.Int31n(20000) % 1234567890
}

func genRandomRect() *rect {
	rand1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	h := GRID_TOP_HEIGHT / (rand1.Int31n(100) + 10)
	rand2 := rand.New(rand.NewSource(time.Now().UnixNano()))
	w := GRID_TOP_WIDTH / (rand2.Int31n(100) + 10)

	lo := genRandomLo()
	la := genRandomLa()
	return &rect{
		lo,
		la,
		lo + w,
		la - h,
	}
}
