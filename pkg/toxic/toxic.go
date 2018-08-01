package toxic

type Analyzer interface {
	ScoreComment(comment string) (float64, error)
}
