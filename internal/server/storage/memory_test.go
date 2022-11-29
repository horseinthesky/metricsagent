package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnsupportedType(t *testing.T) {
	require.True(t, UnsupportedType("unsupported"))
	require.False(t, UnsupportedType("counter"))
}

func TestMemoryDB(t *testing.T) {
	db := NewMemoryStorage()

	ctx := context.Background()

	err := db.Init(ctx)
	require.NoError(t, err)

	err = db.Check(ctx)
	require.NoError(t, err)

	invalid := Metric{
		ID: "invalid",
		MType: "invalid",
	}

	err = db.Set(invalid)
	require.Error(t, err)

	counterValue1 := int64(10)

	counter1 := Metric{
		ID: "testCounter",
		MType: "counter",
		Delta: &counterValue1,
	}

	counterValue2 := int64(10)

	counter2 := Metric{
		ID: "testCounter",
		MType: "counter",
		Delta: &counterValue2,
	}

	counterValue3 := int64(10)

	counter3 := Metric{
		ID: "testCounterNew",
		MType: "counter",
		Delta: &counterValue3,
	}

	gaugeValue := float64(10.0)

	gauge := Metric{
		ID: "testGauge",
		MType: "gauge",
		Value: &gaugeValue,
	}

	metrics := []Metric{counter1, gauge}

	err = db.SetBulk(metrics)
	require.NoError(t, err)

	err = db.SetBulk(metrics)
	require.NoError(t, err)

	err = db.Set(counter1)
	require.NoError(t, err)

	err = db.Set(counter2)
	require.NoError(t, err)

	err = db.Set(counter3)
	require.NoError(t, err)

	err = db.Set(gauge)
	require.NoError(t, err)

	dbCounter, err := db.Get(ctx, "testCounter")
	require.NoError(t, err)
	require.Equal(t, int64(50), *dbCounter.Delta)

	_, err = db.Get(ctx, "notExists")
	require.Error(t, err)

	dbMetrics, err := db.GetAll(ctx)
	require.NoError(t, err)

	dbGauge := dbMetrics["testGauge"]
	require.Equal(t, gaugeValue, *dbGauge.Value)
}
