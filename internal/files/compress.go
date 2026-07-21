package files

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/sunshineplan/imgconv"
)

const (
	cloudinaryCompressAbove = 10 * 1024 * 1024 // compress when larger than 10 MB
	cloudinaryTargetBytes   = 9 * 1024 * 1024  // aim for ~9 MB
)

// prepareCloudinaryUpload returns a path suitable for Cloudinary upload.
// Images larger than 10 MB are re-encoded as JPEG targeting ~9 MB.
// cleanup removes any temporary file created for compression.
func prepareCloudinaryUpload(path string) (uploadPath string, cleanup func(), err error) {
	noop := func() {}

	info, err := os.Stat(path)
	if err != nil {
		return "", noop, err
	}
	if info.Size() <= cloudinaryCompressAbove {
		return path, noop, nil
	}

	log.Default().Printf("[Cloudinary] %s is %.2f MB (>10 MB); compressing to ~9 MB...",
		filepath.Base(path), float64(info.Size())/(1024*1024))

	src, err := imgconv.Open(path)
	if err != nil {
		return "", noop, fmt.Errorf("open image for compression: %w", err)
	}

	data, err := compressImageUnder(src, cloudinaryTargetBytes)
	if err != nil {
		return "", noop, err
	}

	tmp, err := os.CreateTemp("", "film-sync-cld-*.jpg")
	if err != nil {
		return "", noop, fmt.Errorf("create temp jpeg: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", noop, fmt.Errorf("write temp jpeg: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return "", noop, err
	}

	log.Default().Printf("[Cloudinary] Compressed %s -> %.2f MB",
		filepath.Base(path), float64(len(data))/(1024*1024))

	return tmpPath, func() { _ = os.Remove(tmpPath) }, nil
}

// compressImageUnder encodes img as JPEG under targetBytes, lowering quality
// then scaling down if needed.
func compressImageUnder(img image.Image, targetBytes int64) ([]byte, error) {
	current := img
	for scaleAttempt := 0; scaleAttempt < 8; scaleAttempt++ {
		data, ok := encodeJPEGAtOrUnder(current, targetBytes)
		if ok {
			return data, nil
		}
		// Scale to 85% and retry.
		current = imgconv.Resize(current, &imgconv.ResizeOption{Percent: 85})
		log.Default().Printf("[Cloudinary] Still over target; resized to %dx%d",
			current.Bounds().Dx(), current.Bounds().Dy())
	}

	// Last resort: lowest quality after scaling.
	var buf bytes.Buffer
	err := imgconv.Write(&buf, current, &imgconv.FormatOption{
		Format:       imgconv.JPEG,
		EncodeOption: []imgconv.EncodeOption{imgconv.Quality(40)},
	})
	if err != nil {
		return nil, fmt.Errorf("final jpeg encode: %w", err)
	}
	if int64(buf.Len()) > targetBytes {
		return nil, fmt.Errorf("unable to compress image under %d bytes (got %d)", targetBytes, buf.Len())
	}
	return buf.Bytes(), nil
}

func encodeJPEGAtOrUnder(img image.Image, targetBytes int64) ([]byte, bool) {
	low, high := 40, 92
	var best []byte

	for low <= high {
		q := (low + high) / 2
		var buf bytes.Buffer
		err := imgconv.Write(&buf, img, &imgconv.FormatOption{
			Format:       imgconv.JPEG,
			EncodeOption: []imgconv.EncodeOption{imgconv.Quality(q)},
		})
		if err != nil {
			return nil, false
		}
		if int64(buf.Len()) <= targetBytes {
			best = buf.Bytes()
			low = q + 1 // try higher quality
		} else {
			high = q - 1
		}
	}

	return best, best != nil
}
