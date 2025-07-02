package cbreaker

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/stretchr/testify/require"
)

func TestEngine_IsOpen(t *testing.T) {
	t.Run(
		"is open is false in closed state", func(t *testing.T) {
			e := &Engine{
				state: ClosedState,
			}

			require.False(t, e.IsOpen())
		},
	)

	t.Run(
		"is open is false in half open state", func(t *testing.T) {
			e := &Engine{
				state: HalfOpenState,
			}

			require.False(t, e.IsOpen())
		},
	)

	t.Run(
		"is open is true in  open state", func(t *testing.T) {
			e := &Engine{
				state: OpenState,
			}

			require.True(t, e.IsOpen())
		},
	)

	t.Run(
		"close to close state with success", func(t *testing.T) {
			e := &Engine{
				state:              ClosedState,
				consecutiveFailure: 10,
			}

			e.ReportSuccess()

			require.Equal(t, &Engine{
				state: ClosedState,
			}, e)
		},
	)

	t.Run(
		"close to close state with failure", func(t *testing.T) {
			e := &Engine{
				state: ClosedState,
			}

			e.ReportFailure()

			require.Equal(t, &Engine{
				state:              ClosedState,
				consecutiveFailure: 1,
			}, e)
		},
	)

	t.Run(
		"close to open state with success", func(t *testing.T) {
			e := &Engine{
				state: ClosedState,
			}

			e.ReportFailure()

			require.Equal(t, &Engine{
				state:              ClosedState,
				consecutiveFailure: 1,
			}, e)
		},
	)

	t.Run(
		"close to open with failure", func(t *testing.T) {
			c := clock.NewMock()

			e := &Engine{
				config: config{
					halfOpenTimeout: ref(time.Second),
					nbFailureToOpen: ref(4),
				},
				state:              ClosedState,
				consecutiveFailure: 3,
				clock:              c,
			}

			e.ReportFailure()

			require.Equal(t, &Engine{
				config: config{
					halfOpenTimeout: ref(time.Second),
					nbFailureToOpen: ref(4),
				},
				state: OpenState,
				clock: c,
			}, e)

			time.Sleep(time.Millisecond)

			c.Add(20 * time.Second)

			require.Equal(t, &Engine{
				config: config{
					halfOpenTimeout: ref(time.Second),
					nbFailureToOpen: ref(4),
				},
				state: HalfOpenState,
				clock: c,
			}, e)

		},
	)

	t.Run(
		"half open with failure", func(t *testing.T) {
			c := clock.NewMock()
			e := &Engine{
				state: HalfOpenState,
				clock: c,
			}

			e.ReportFailure()

			require.Equal(t, &Engine{
				state: OpenState,
				clock: c,
			}, e)
		},
	)

	t.Run(
		"half open to half open with success", func(t *testing.T) {
			c := clock.NewMock()

			e := &Engine{
				config: config{
					nbSuccessToClose: ref(2),
				},
				state: HalfOpenState,
				clock: c,
			}

			e.ReportSuccess()

			require.Equal(t, &Engine{
				config: config{
					nbSuccessToClose: ref(2),
				},
				state:              HalfOpenState,
				consecutiveSuccess: 1,
				clock:              c,
			}, e)
		},
	)

	t.Run(
		"half open to closed state with success", func(t *testing.T) {
			c := clock.NewMock()

			e := &Engine{
				config: config{
					nbSuccessToClose: ref(2),
				},
				consecutiveSuccess: 1,
				state:              HalfOpenState,
				clock:              c,
			}

			e.ReportSuccess()

			require.Equal(t, &Engine{
				config: config{
					nbSuccessToClose: ref(2),
				},
				state: ClosedState,
				clock: c,
			}, e)
		},
	)
}

func Test_toto(t *testing.T) {
	e, err := newEngine(clock.New(),
		WithOpenTimeout(time.Second),
		WithFailuresToOpen(3),
		WithSuccessToClose(3),
	)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)

		if e.IsOpen() {
			fmt.Println(
				"       ",
				e.IsOpen(),
				e.state,
				e.consecutiveSuccess,
				e.consecutiveFailure,
			)

			continue
		}

		if rand.Float64() < 0.2 {
			e.ReportFailure()
			fmt.Println(
				"failure",
				e.IsOpen(),
				e.state,
				e.consecutiveSuccess,
				e.consecutiveFailure,
			)

			continue
		}

		e.ReportSuccess()
		fmt.Println(
			"success",
			e.IsOpen(),
			e.state,
			e.consecutiveSuccess,
			e.consecutiveFailure,
		)
	}
}
