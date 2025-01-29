package indexer

type Indexer interface {
	Run()
	Name() string
}
