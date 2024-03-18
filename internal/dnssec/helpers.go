package dnssec

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/exp/constraints"
)

func nameClassTypeToString(qname string, qClass, qType uint16) string {
	return qname + " " + dns.ClassToString[qClass] + " " + dns.TypeToString[qType]
}

func nameTypeToString(qname string, qType uint16) string {
	return qname + " " + dns.TypeToString[qType]
}

func hashToString(hashType uint8) string {
	s, ok := dns.HashToString[hashType]
	if ok {
		return s
	}
	return fmt.Sprintf("%d", hashType)
}

func hashesToString(hashTypes []uint8) string {
	hashStrings := make([]string, len(hashTypes))
	for i, hash := range hashTypes {
		hashStrings[i] = hashToString(hash)
	}
	return strings.Join(hashStrings, ", ")
}

func integersToString[T constraints.Integer](integers []T) string {
	integerStrings := make([]string, len(integers))
	for i, hash := range integers {
		integerStrings[i] = fmt.Sprint(hash)
	}
	return strings.Join(integerStrings, ", ")
}

func wrapError(zone string, qClass, qType uint16, err error) error {
	return fmt.Errorf("for %s: %w", nameClassTypeToString(zone, qClass, qType), err)
}

func isOneOf[T comparable](value T, values ...T) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}
