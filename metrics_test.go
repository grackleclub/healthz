package healthz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	load1, load5, load15, err := Load()
	require.NoError(t, err)
	t.Logf("Load: %f %f %f", load1, load5, load15)
}

func TestCPU(t *testing.T) {
	cpu, err := CPU()
	require.NoError(t, err)
	t.Logf("CPU: %f", cpu)
}

func TestMEM(t *testing.T) {
	memory, err := MEM()
	require.NoError(t, err)
	t.Logf("Memory: %f", memory)
}

func TestDISK(t *testing.T) {
	disk, err := DISK()
	require.NoError(t, err)
	t.Logf("Disk: %f", disk)
}
