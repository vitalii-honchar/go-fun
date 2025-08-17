package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Eyevinn/mp4ff/mp4"
)

func main() {
	fileName := "/Users/vitaliihonchar/Documents/OBS/2025-08-14 07-30-14.mp4"
	
	err := extractAudio(fileName)
	if err != nil {
		log.Fatal("Error extracting audio:", err)
	}
}

func extractAudio(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	mp4File, err := mp4.DecodeFile(file)
	if err != nil {
		return fmt.Errorf("failed to decode MP4 file: %w", err)
	}

	fmt.Printf("Total tracks found: %d\n", len(mp4File.Moov.Traks))
	
	if len(mp4File.Segments) > 0 {
		fmt.Printf("Found %d segments (fragmented MP4)\n", len(mp4File.Segments))
		
		audioTrackID := uint32(0)
		var audioCodec string
		var fileExtension string
		
		for _, track := range mp4File.Moov.Traks {
			if track.Mdia.Hdlr.HandlerType == "soun" {
				audioTrackID = track.Tkhd.TrackID
				
				if track.Mdia.Minf.Stbl.Stsd != nil && len(track.Mdia.Minf.Stbl.Stsd.Children) > 0 {
					switch track.Mdia.Minf.Stbl.Stsd.Children[0].Type() {
					case "mp4a":
						audioCodec = "AAC"
						fileExtension = ".aac"
					case "mp3 ":
						audioCodec = "MP3"
						fileExtension = ".mp3"
					default:
						audioCodec = track.Mdia.Minf.Stbl.Stsd.Children[0].Type()
						fileExtension = ".audio"
					}
				} else {
					audioCodec = "Unknown"
					fileExtension = ".audio"
				}
				
				fmt.Printf("Audio track ID: %d, Codec: %s\n", audioTrackID, audioCodec)
				break
			}
		}
		
		if audioTrackID > 0 {
			outputFileName := "out" + fileExtension
			outputFile, err := os.Create(outputFileName)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outputFile.Close()
			
			totalSamples := 0
			totalBytes := 0
			
			fmt.Printf("Extracting audio samples to %s...\n", outputFileName)
			
			for _, segment := range mp4File.Segments {
				for _, fragment := range segment.Fragments {
					if fragment.Moof != nil && fragment.Mdat != nil {
						for _, traf := range fragment.Moof.Trafs {
							if traf.Tfhd.TrackID == audioTrackID && traf.Trun != nil {
								sampleCount := int(traf.Trun.SampleCount())
								
								offset := uint64(0)
								for i := 0; i < sampleCount; i++ {
									sampleSize := traf.Trun.Samples[i].Size
									sampleData := fragment.Mdat.Data[offset : offset+uint64(sampleSize)]
									
									if audioCodec == "AAC" {
										// Add ADTS header for each AAC frame
										adtsHeader := createADTSHeader(int(sampleSize))
										_, err := outputFile.Write(adtsHeader)
										if err != nil {
											return fmt.Errorf("failed to write ADTS header: %w", err)
										}
										totalBytes += len(adtsHeader)
									}
									
									bytesWritten, err := outputFile.Write(sampleData)
									if err != nil {
										return fmt.Errorf("failed to write sample data: %w", err)
									}
									
									totalBytes += bytesWritten
									offset += uint64(sampleSize)
								}
								
								totalSamples += sampleCount
							}
						}
					}
				}
			}
			
			fmt.Printf("Extracted %d audio samples (%d bytes) to %s\n", totalSamples, totalBytes, outputFileName)
		}
	} else {
		for i, track := range mp4File.Moov.Traks {
			fmt.Printf("Track %d: Handler=%s, TrackID=%d\n", i+1, track.Mdia.Hdlr.HandlerType, track.Tkhd.TrackID)
			
			if track.Mdia.Hdlr.HandlerType == "soun" {
				fmt.Printf("Found audio track: ID=%d\n", track.Tkhd.TrackID)
				
				stbl := track.Mdia.Minf.Stbl
				
				if stbl.Stsz != nil {
					fmt.Printf("Sample count (Stsz): %d\n", stbl.Stsz.SampleNumber)
				}
			}
		}
	}

	return nil
}

func createADTSHeader(frameLength int) []byte {
	header := make([]byte, 7)
	
	// ADTS fixed header
	header[0] = 0xFF
	header[1] = 0xF1 // MPEG-4, no CRC
	header[2] = 0x50 // AAC LC, 44.1kHz, mono
	
	// Frame length (13 bits) + other fields
	frameLen := frameLength + 7 // Add ADTS header length
	header[3] = byte((frameLen >> 11) & 0x03)
	header[4] = byte((frameLen >> 3) & 0xFF)
	header[5] = byte(((frameLen & 0x07) << 5) | 0x1F)
	header[6] = 0xFC
	
	return header
}
