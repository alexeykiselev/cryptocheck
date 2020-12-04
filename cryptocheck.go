package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/alexeykiselev/cryptocheck/internal"
	"github.com/vardius/worker-pool/v2"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const recordSize = 4 + 4 + 64

type record struct {
	n   uint64
	l   uint64
	sig [64]byte
}

func (r *record) unmarshal(data []byte) error {
	if l := len(data); l < recordSize {
		return fmt.Errorf("insufficient data size %d", l)
	}
	r.n = uint64(binary.BigEndian.Uint32(data[0:4]))
	r.l = uint64(binary.BigEndian.Uint32(data[4:8]))
	copy(r.sig[:], data[8:])
	return nil
}

func (r *record) String() string {
	return fmt.Sprintf("[%d, %d, %s]", r.n, r.l, hex.EncodeToString(r.sig[:]))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()
	err := run(ctx)
	if err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	p := message.NewPrinter(language.English)
	start := time.Now()
	var filename string
	flag.StringVar(&filename, "filename", "", "Test parameters file")
	flag.Parse()

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()
	if size < recordSize {
		return errors.New(fmt.Sprintf("file too small %d", size))
	}

	seed := make([]byte, 8)
	_, err = f.Read(seed)
	if err != nil {
		return err
	}
	seedNumber := binary.BigEndian.Uint64(seed)
	template := internal.Template(seed)
	log.Printf("Seed: %s(%d)", hex.EncodeToString(seed), seedNumber)
	size = size - 8
	count := size / recordSize
	log.Printf("Records count: %s", p.Sprintf("%d", count))
	if size%recordSize != 0 {
		return errors.New(fmt.Sprintf("invalid file size %d", size))
	}
	workers := runtime.NumCPU()
	pool := workerpool.New(int(count) / workers)

	report := make(chan struct{}, workers)
	failures := make(chan error, workers)

	worker := func(r record) {
		defer func() {
			report <- struct{}{}
		}()
		as := internal.AccountSeed(seedNumber, r.n)
		sk := crypto.GenerateSecretKey(as)
		pk := crypto.GeneratePublicKey(sk)
		if err != nil {
			failures <- err
			return
		}
		msg := internal.Message(template, r.l)
		if !crypto.Verify(pk, r.sig, msg) {
			failures <- fmt.Errorf("invalid signature for record %s, message %s, pk %s, ",
				r.String(), hex.EncodeToString(msg), pk.String())
			return
		}
	}

	for i := 0; i < workers; i++ {
		if err := pool.AddWorker(worker); err != nil {
			return err
		}
	}
	defer pool.Stop()

	go readRecords(ctx, bufio.NewReader(f), pool, failures)

	go func() {
		for f := range failures {
			log.Printf("CHECK FAILED: %v", f)
			os.Exit(1)
		}
	}()

	var done int64 = 0
	for done < count {
		select {
		case <-report:
			done++
		case <-ctx.Done():
			return fmt.Errorf("user termination")
		}
		if done%100000 == 0 {
			log.Printf("%s records checked", p.Sprintf("%d", done))
		}

	}
	log.Printf("DONE %s recods in %s", p.Sprintf("%d", done), time.Since(start))

	return nil
}

func readRecords(ctx context.Context, r io.Reader, pool workerpool.Pool, failures chan error) {
	buf := make([]byte, recordSize)
	for {
		select {
		case <-ctx.Done():
			failures <- fmt.Errorf("user termination")
			return
		default:
			_, err := io.ReadFull(r, buf)
			if err != nil {
				if err != io.EOF {
					failures <- fmt.Errorf("unable to read record: %v", err)
					return
				}
				return
			}
			r := record{}
			err = r.unmarshal(buf)
			if err != nil {
				failures <- fmt.Errorf("unable to unmarshal record: %v", err)
				return
			}
			err = pool.Delegate(r)
			if err != nil {
				failures <- err
				return
			}
		}
	}
}
