package main

func sideWays(inbox box) (outbox box) {
	outbox = inbox
	if outbox.w < outbox.l {
		outbox.l, outbox.w = outbox.w, outbox.l
	}
	return
}

func upRight(inbox box) (outbox box) {
	outbox = inbox
	if outbox.l < outbox.w {
		outbox.l, outbox.w = outbox.w, outbox.l
	}
	return
}
