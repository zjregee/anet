package anet

import "unsafe"

func (p *defaultRing) getOperator(fd int, ptr unsafe.Pointer) *FDOperator {
	return *(**FDOperator)(ptr)
}

func (p *defaultRing) setOperator(ptr unsafe.Pointer, operator *FDOperator) {
	*(**FDOperator)(ptr) = operator
}
