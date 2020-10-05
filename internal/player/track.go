package player

import (
	"io"
	"sync"
	"time"

	"github.com/kvark128/OnlineLibrary/internal/winmm"
	"github.com/kvark128/minimp3"
	"github.com/kvark128/sonic"
)

type track struct {
	sync.Mutex
	stopped     bool
	paused      bool
	beRewind    bool
	elapsedTime time.Duration
	start       time.Time
	stream      *sonic.Stream
	dec         *minimp3.Decoder
	sampleRate  int
	channels    int
	samples     []byte
	wp          *winmm.WavePlayer
}

func newTrack(mp3 io.Reader, speed float64) *track {
	trk := &track{}
	trk.dec = minimp3.NewDecoder(mp3)
	trk.dec.Read([]byte{}) // Reads first frame
	trk.sampleRate, trk.channels, _, _ = trk.dec.Info()
	trk.stream = sonic.NewStream(trk.sampleRate, trk.channels)
	trk.stream.SetSpeed(speed)
	trk.samples = make([]byte, trk.sampleRate*trk.channels*2) // buffer for 1 second
	trk.wp = winmm.NewWavePlayer(trk.channels, trk.sampleRate, 16, len(trk.samples), winmm.WAVE_MAPPER)
	return trk
}

func (trk *track) play() {
	trk.start = time.Now()
	for {
		trk.Lock()
		if trk.stopped {
			trk.Unlock()
			break
		}
		n, err := trk.dec.Read(trk.samples)
		trk.stream.Write(trk.samples[:n])
		if err != nil {
			trk.stream.Flush()
		}
		trk.Unlock()
		for {
			trk.Lock()
			if trk.stopped {
				trk.Unlock()
				break
			}
			n, _ := trk.stream.Read(trk.samples)
			trk.Unlock()
			if n == 0 {
				break
			}
			trk.wp.Write(trk.samples[:n])
		}
		if err != nil {
			break
		}
	}

	trk.wp.Sync()
	trk.wp.Close()
	trk.stream.Close()
}

func (trk *track) setSpeed(speed float64) {
	trk.Lock()
	if !trk.paused {
		trk.elapsedTime += time.Duration(float64(time.Since(trk.start)) * trk.stream.Speed())
		trk.start = time.Now()
	}
	trk.stream.SetSpeed(speed)
	trk.Unlock()
}

func (trk *track) getElapsedTime() time.Duration {
	trk.Lock()
	defer trk.Unlock()
	if trk.paused {
		return trk.elapsedTime
	}
	return trk.elapsedTime + time.Duration(float64(time.Since(trk.start))*trk.stream.Speed())
}

func (trk *track) pause(pause bool) bool {
	trk.Lock()
	defer trk.Unlock()

	if trk.paused == pause {
		return false
	}
	trk.paused = pause

	if trk.paused {
		if !trk.start.IsZero() {
			trk.elapsedTime += time.Duration(float64(time.Since(trk.start)) * trk.stream.Speed())
		}
	} else {
		trk.start = time.Now()
	}
	if trk.beRewind {
		trk.beRewind = false
		trk.wp.Stop()
		return true
	}
	trk.wp.Pause(trk.paused)
	return true
}

func (trk *track) stop() {
	trk.Lock()
	trk.stopped = true
	trk.wp.Stop()
	trk.Unlock()
}

func (trk *track) rewind(offset time.Duration) error {
	if trk.pause(true) {
		defer trk.pause(false)
	}

	trk.Lock()
	defer trk.Unlock()

	elapsedTime := trk.elapsedTime + offset
	if elapsedTime < 0 {
		elapsedTime = time.Duration(0)
	}
	offsetInBytes := int64(float64(trk.sampleRate*trk.channels*2) * elapsedTime.Seconds())

	if _, err := trk.dec.Seek(offsetInBytes, io.SeekStart); err != nil {
		return err
	}

	for {
		nSamples, _ := trk.stream.Read(trk.samples)
		if nSamples == 0 {
			break
		}
	}

	trk.elapsedTime = elapsedTime
	trk.beRewind = true
	return nil
}
