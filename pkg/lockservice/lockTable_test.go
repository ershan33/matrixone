// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lockservice

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRowLock(t *testing.T) {
	l := NewLockService()
	ctx := context.Background()
	option := LockOptions{
		granularity: Row,
		mode:        Exclusive,
		policy:      Wait,
	}
	acquired := false

	ok, err := l.Lock(context.Background(), 0, [][]byte{{1}}, []byte{1}, option)
	assert.NoError(t, err)
	assert.Equal(t, true, ok)
	go func() {
		ok, err := l.Lock(ctx, 0, [][]byte{{1}}, []byte{2}, option)
		assert.NoError(t, err)
		assert.Equal(t, true, ok)
		acquired = true
		err = l.Unlock([]byte{2})
		assert.NoError(t, err)
	}()
	time.Sleep(time.Second)
	err = l.Unlock([]byte{1})
	assert.NoError(t, err)
	time.Sleep(time.Second)
	ok, err = l.Lock(context.Background(), 0, [][]byte{{1}}, []byte{3}, option)
	assert.NoError(t, err)
	assert.Equal(t, true, ok)

	assert.Equal(t, true, acquired)

	err = l.Unlock([]byte{3})
	assert.NoError(t, err)
}

func TestMultipleRowLocks(t *testing.T) {
	l := NewLockService()
	ctx := context.Background()
	option := LockOptions{
		granularity: Row,
		mode:        Exclusive,
		policy:      Wait,
	}
	iter := 0
	sum := 1000
	var wg sync.WaitGroup

	for i := 0; i < sum; i++ {
		wg.Add(1)
		go func(i int) {
			ok, err := l.Lock(ctx, 0, [][]byte{{1}}, []byte(strconv.Itoa(i)), option)
			assert.NoError(t, err)
			assert.Equal(t, true, ok)
			iter++
			err = l.Unlock([]byte(strconv.Itoa(i)))
			assert.NoError(t, err)
			wg.Done()
		}(i)
	}
	wg.Wait()
	assert.Equal(t, sum, iter)
}

func TestRangeLock(t *testing.T) {
	l := NewLockService()
	ctx := context.Background()
	option := LockOptions{
		granularity: Range,
		mode:        Exclusive,
		policy:      Wait,
	}
	acquired := false

	ok, err := l.Lock(context.Background(), 0, [][]byte{{1}, {2}}, []byte{1}, option)
	assert.NoError(t, err)
	assert.Equal(t, true, ok)
	go func() {
		ok, err := l.Lock(ctx, 0, [][]byte{{1}, {2}}, []byte{2}, option)
		assert.NoError(t, err)
		assert.Equal(t, true, ok)
		acquired = true
		err = l.Unlock([]byte{2})
		assert.NoError(t, err)
	}()
	time.Sleep(time.Second)
	err = l.Unlock([]byte{1})
	assert.NoError(t, err)
	time.Sleep(time.Second)
	ok, err = l.Lock(context.Background(), 0, [][]byte{{1}, {2}}, []byte{3}, option)
	assert.NoError(t, err)
	assert.Equal(t, true, ok)

	assert.Equal(t, true, acquired)

	err = l.Unlock([]byte{3})
	assert.NoError(t, err)
}

func TestMultipleRangeLocks(t *testing.T) {
	l := NewLockService()
	ctx := context.Background()
	option := LockOptions{
		granularity: Range,
		mode:        Exclusive,
		policy:      Wait,
	}

	sum := 100
	var wg sync.WaitGroup
	for i := 0; i < sum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			start := i % 10
			if start == 9 {
				return
			}
			end := (i + 1) % 10
			ok, err := l.Lock(ctx, 0, [][]byte{{byte(start)}, {byte(end)}}, []byte(strconv.Itoa(i)), option)
			assert.NoError(t, err)
			assert.Equal(t, true, ok)
			err = l.Unlock([]byte(strconv.Itoa(i)))
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()
}

func BenchmarkMultipleRowLock(b *testing.B) {
	l := NewLockService()
	ctx := context.Background()
	option := LockOptions{
		granularity: Row,
		mode:        Exclusive,
		policy:      Wait,
	}
	iter := 0

	b.Run("lock-service", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			go func(i int) {
				l.Lock(ctx, 0, [][]byte{{1}}, []byte{byte(i)}, option)
				iter++
				l.Unlock([]byte{byte(i)})
			}(i)
		}
	})
}
