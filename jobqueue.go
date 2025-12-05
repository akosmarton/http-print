package main

type Job struct {
	ContentType string
	Payload     []byte
}

type JobQueue struct {
	printers map[string]chan Job
	len      int
}

func (q *JobQueue) Get(printer string) chan Job {
	return q.printers[printer]
}

func (q *JobQueue) Init(printer string) {
	q.printers[printer] = make(chan Job, q.len)
}

func (q *JobQueue) Delete(printer string) {
	delete(q.printers, printer)
}

func (q *JobQueue) Clear(printer string) {
	q.printers[printer] = make(chan Job, q.len)
}

func NewJobQueue(len int) *JobQueue {
	return &JobQueue{
		printers: make(map[string]chan Job),
		len:      len,
	}
}
