//go:generate easyjson -all stats.go

package hw10programoptimization

import (
	"bufio"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	return getStat(r, domain)
}

func getStat(r io.Reader, domain string) (result DomainStat, err error) {
	scanner := bufio.NewScanner(r)
	result = make(DomainStat)

	user := &User{}
	for scanner.Scan() {
		if err = user.UnmarshalJSON(scanner.Bytes()); err != nil {
			return
		}
		if strings.Contains(user.Email, "."+domain) {
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]++
		}
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}
