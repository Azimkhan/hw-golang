package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func stageWrapper(stage Stage, in In, done In) Out {
	newIn := make(Bi)
	go func() {
		defer close(newIn)
		for {
			// prioritize done
			select {
			case <-done:
				return
			default:
			}

			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				newIn <- v
			}
		}
	}()
	return stage(newIn)
}

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	nextIn := in
	for _, stage := range stages {
		nextIn = stageWrapper(stage, nextIn, done)
	}
	return nextIn
}
