package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func pipeChan(externalCh In, done In, pipeCh Bi) {
	defer close(pipeCh)

	for {
		select {
		case <-done:
			go func() {
				for range externalCh {
					// вычитываем данные последней стадии, чтобы она закрылась и закрыла цепочку предыдущих стадий.
					continue
				}
			}()
			return
		case v, ok := <-externalCh:
			if !ok {
				return
			}
			select {
			case <-done:
				go func() {
					for range externalCh {
						// вычитываем данные последней стадии, чтобы она закрылась и закрыла цепочку предыдущих стадий.
						continue
					}
				}()
				return
			case pipeCh <- v:
			}
		}
	}
}

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	pipeIn := make(Bi) // in неизвестный канал. сделаем свой для входящих значений.

	// запустим горутину, которая мониторит done и пишет в pipeIn.
	go pipeChan(in, done, pipeIn)

	stageIn := In(pipeIn)
	for _, gStage := range stages {
		// каждай горутина-стедж создаст свой выходной канал и передаст его в следующий стейдж.
		// таким образом стеджи будут выполняться последовательно для каждого значения из in (обрабатываемых параллельно).
		stageIn = gStage(stageIn) // (gStage(stageIn) = sOutN).
	}

	resCh := make(Bi)

	// запустим горутину, которая мониторит done и вычитывает канал stageIn в канал resCh.
	go pipeChan(stageIn, done, resCh)

	return Out(resCh)
}
