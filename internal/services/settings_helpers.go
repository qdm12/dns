package services

import (
	"fmt"
	"sort"
)

func validateServicesAreUnique(services []Service) (errMessage string) {
	duplicateServices, duplicatedNames := findDuplicatedServices(services)

	if len(duplicateServices) > 0 {
		return makeDuplicatedServicesErrMessage(duplicateServices, "service")
	}

	if len(duplicatedNames) == 0 {
		return ""
	}

	// Convert map[string]uint to map[fmt.Stringer]uint
	// for service name strings
	duplicatedNameStringers := make(map[fmt.Stringer]uint, len(duplicatedNames))
	for serviceString, count := range duplicatedNames {
		nameStringer := &stringer{s: serviceString}
		duplicatedNameStringers[nameStringer] = count
	}

	return makeDuplicatedServicesErrMessage(duplicatedNameStringers, "service name")
}

type stringer struct{ s string }

func (s *stringer) String() string { return s.s }

func findDuplicatedServices(services []Service) (
	duplicatedServices map[fmt.Stringer]uint,
	duplicatedNames map[string]uint) {
	duplicatedServices = make(map[fmt.Stringer]uint, len(services))
	duplicatedNames = make(map[string]uint, len(services))
	for _, service := range services {
		duplicatedServices[service]++
		duplicatedNames[service.String()]++
	}

	for service, count := range duplicatedServices {
		if count == 1 {
			delete(duplicatedServices, service)
		}

		serviceString := service.String()
		if duplicatedNames[serviceString] == 1 {
			delete(duplicatedNames, serviceString)
		}
	}

	return duplicatedServices, duplicatedNames
}

func makeDuplicatedServicesErrMessage(duplicatedServices map[fmt.Stringer]uint,
	messagePrefix string) (errMessage string) {
	switch len(duplicatedServices) {
	case 0:
		return ""
	case 1:
		var service fmt.Stringer
		var count uint
		for service, count = range duplicatedServices {
			break
		}
		return fmt.Sprintf("%s %s is duplicated %s",
			messagePrefix, service, countToString(count))
	default:
		parts := make(sort.StringSlice, 0, len(duplicatedServices))
		for service, count := range duplicatedServices {
			part := fmt.Sprintf("%s is duplicated %s",
				service, countToString(count))
			parts = append(parts, part)
		}
		parts.Sort() // predictable order for tests
		return messagePrefix + "s " + andStrings(parts)
	}
}

func countToString(count uint) string {
	const zero, one, two = 0, 1, 2
	switch count {
	case zero:
		return "0 time"
	case one:
		return "once"
	case two:
		return "twice"
	default:
		return fmt.Sprintf("%d times", count)
	}
}
