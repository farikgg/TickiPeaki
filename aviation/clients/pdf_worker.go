package clients

import "aviation/models"

func StartPDFWorker(
	client *PDFClient,
	ticket models.Ticket,
	onSuccess func(pdfURL string),
	onFailure func(err error),
) {
	go func() {
		if err := client.RequestGeneration(ticket); err != nil {
			onFailure(err)
			return
		}

		url, err := client.PollStatus(ticket.ID)
		if err != nil {
			onFailure(err)
			return
		}

		onSuccess(*url)
	}()
}
