package main

import (
	"fmt"
	"time"
)

type State int

const (
	UnknownState State = iota
	Success
	Failure
)

type Counter interface {
	Count(State)
	ConsecutiveFailures() int
	LastActivity() time.Time
	Reset()
}

type Count struct {
	CurrentState State
	Failure      int
	Success      int
	CFailure     int
	CSuccess     int
	TimeStamp    time.Time
}

func (a *Count) LastActivity() time.Time {
	return a.TimeStamp
}

func (a *Count) ConsecutiveFailures() int {
	return a.CFailure
}

func (a *Count) Reset() {
	*a = Count{
		TimeStamp: time.Now(),
	}
}

func (a *Count) Count(s State) {
	switch s {
	case Success:
		a.Success++
		if a.CurrentState == Success {
			a.CFailure = 0
			a.Failure--
			a.CSuccess++
		}
		a.CurrentState = Success
	case Failure:
		a.Failure++
		if a.CurrentState == Failure {
			a.CSuccess = 0
			a.Failure--
			a.CFailure++
		}
		a.CurrentState = Failure
	}
}

func newCounter() Counter {
	return &Count{
		TimeStamp: time.Now(),
	}
}

type FunctionToExecute func(interface{}) (interface{}, error)

// Closure
func Breaker(f FunctionToExecute, failureCount int) FunctionToExecute {
	c := newCounter()

	return func(ctx interface{}) (interface{}, error) {
		if c.ConsecutiveFailures() >= failureCount {
			canRetry := func(c Counter) bool {
				backoffLevel := c.ConsecutiveFailures() - failureCount

				shouldRetry := c.LastActivity().Add(time.Second * 2 << backoffLevel)
				return time.Now().After(shouldRetry)
			}

			if !canRetry(c) {
				return nil, fmt.Errorf("Service Unavailable")
			}
		}

		resp, err := f(ctx)
		if err != nil {
			c.Count(Failure)
			return nil, err
		}

		c.Count(Success)
		return resp, nil
	}
}

func test(k interface{}) (interface{}, error) {
	if k.(string) == "Anton" {
		return nil, fmt.Errorf("Error")
	}
	return fmt.Sprintf("Hello %s", k.(string)), nil
}

func main() {
	testwithwrappedcircuitbreaker := Breaker(test, 2)
	var err error
	_, err = testwithwrappedcircuitbreaker("Anton")
	fmt.Println(err)
	_, err = testwithwrappedcircuitbreaker("Anton")
	fmt.Println(err)
	_, err = testwithwrappedcircuitbreaker("Anton")
	fmt.Println(err)
	_, err = testwithwrappedcircuitbreaker("Anton")
	fmt.Println(err)
	_, err = testwithwrappedcircuitbreaker("Anton")
	fmt.Println(err)
	_, err = testwithwrappedcircuitbreaker("Anton")
	fmt.Println(err)

}
