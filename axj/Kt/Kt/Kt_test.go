package Kt

import (
	"axj/Kt/KtUnsafe"
	"testing"
)

type A struct {
	a string
}

func (a *A) say1() {
	println("a1")
	a.say2()
}

func (a *A) say2() {
	println("a2")
}

type B struct {
	A
	b string
}

func (b *B) say2() {
	println("b2")
}

func TestUnsafe(t *testing.T) {
	b := B{}
	b.say1()
	b.say2()
	b.A.say1()

	b.b = "ggg"

	println(KtUnsafe.PointerHash(b))
	println(KtUnsafe.PointerHash(&b))
	println(KtUnsafe.PointerHash(new(B)))
	println(KtUnsafe.PointerHash(new(B)))
	println(KtUnsafe.PointerHash(b))
	println(KtUnsafe.PointerHash(&b))
}
