package models

type AckPrintable interface {
	AckPrint(fromClient bool)
}
