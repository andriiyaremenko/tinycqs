package query

import "sync"

func NewQueryResultWriter() ResultWriter {
	return &queryResultReadWriter{ch: make(chan Result)}
}

type queryResultReadWriter struct {
	isDone bool
	ch     chan Result

	rwMu    sync.RWMutex
	once    sync.Once
	writeWG sync.WaitGroup
}

func (rw *queryResultReadWriter) Read() <-chan Result {
	return rw.ch
}

func (rw *queryResultReadWriter) Write(queryResult Result) {
	rw.writeWG.Add(1)
	go func() {
		rw.rwMu.RLock()
		defer rw.rwMu.RUnlock()
		defer rw.writeWG.Done()

		if !rw.isDone {
			rw.ch <- queryResult
		}
	}()
}

func (rw *queryResultReadWriter) Done() {
	rw.once.Do(func() {
		rw.writeWG.Wait()
		rw.rwMu.Lock()

		rw.isDone = true

		close(rw.ch)
		rw.rwMu.Unlock()
	})
}

func (rw *queryResultReadWriter) GetReader() ResultReader {
	return rw
}
