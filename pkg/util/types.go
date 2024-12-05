package utils

type FloatType interface {
	float32 | float64
}

type IntegerType interface {
	int8 | int16 | int32 | int64 | int |
		uint8 | uint16 | uint32 | uint64 | uint
}

type ComplexType interface {
	complex64 | complex128
}

type Number interface {
	FloatType | IntegerType | ComplexType
}
