package utils

import "fmt"

func formatSize(size int64) string {
	const (
		byteSize = 1.0
		kiloSize = 1024.0
		megaSize = 1024.0 * kiloSize
		gigaSize = 1024.0 * megaSize
	)

	switch {
	case size < kiloSize:
		return fmt.Sprintf("%d B", size)
	case size < megaSize:
		return fmt.Sprintf("%.2f KB", float64(size)/kiloSize)
	case size < gigaSize:
		return fmt.Sprintf("%.2f MB", float64(size)/megaSize)
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/gigaSize)
	}
}
