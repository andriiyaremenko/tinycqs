package query

import "sync"

func NewQueryResultWriter() QueryResultWriter {
	return &queryResultReadWriter{ch: make(chan QueryResult)}
}

type queryResultReadWriter struct {
	isDone bool
	ch     chan QueryResult

	rwMu    sync.RWMutex
	once    sync.Once
	writeWG sync.WaitGroup
}

func (rw *queryResultReadWriter) Read() <-chan QueryResult {
	return rw.ch
}

func (rw *queryResultReadWriter) Write(queryResult QueryResult) {
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

func (rw *queryResultReadWriter) GetReader() QueryResultReader {
	return rw
}
