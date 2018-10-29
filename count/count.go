package count

import (
	"context"
	"io"
	"math/big"
)

const averageWebpageSize = 2e+6

func countSubstring(ctx context.Context, source io.Reader, buf, substr []byte) (total *big.Int, err error) {
	total = big.NewInt(0)
	substrIndex := 0
	substrLen := len(substr)
	if substrLen == 0 {
		panic("length of substring must be bigger than zero")
	}

	var subtotal int64
	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
			n, err := source.Read(buf)
			subtotal = 0
			for bufIndex := 0; bufIndex < n; bufIndex++ {
				if buf[bufIndex] == substr[substrIndex] {
					substrIndex++
				} else {
					substrIndex = 0
					if buf[bufIndex] == substr[substrIndex] {
						substrIndex++
					}
				}
				if substrIndex == substrLen {
					subtotal++
					substrIndex = 0
				}
			}

			total.Add(total, big.NewInt(subtotal))

			if err != nil {
				if err == io.EOF {
					err = nil
				}
				return total, err
			}
		}

	}
	return
}

func processSource(ctx context.Context, source source, buf, substring []byte) (subtotal *big.Int, err error) {
	defer source.Close()
	if err = source.load(ctx); err != nil {
		return
	}

	return countSubstring(ctx, source, buf, substring)
}
