package ex

import "golang.org/x/exp/slog"

func Handle(err error) {
	slog.Error("app got error", slog.String("error msg", err.Error()))
	panic(err)
}
