package tts

import (
	"context"
	"fmt"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

// GoogleTextToSpeech converts the given text to speech, saves to the specified filePath
func GoogleTextToSpeech(text, filePath string, language string) error {

   ctx := context.Background()

    // Create a new client
    client, err := texttospeech.NewClient(ctx)
    if err != nil {
        return fmt.Errorf("failed to create TTS client: %w", err)
    }
    defer client.Close()

    var TTSCode string
    var TTSName string

    if language == "german" {
        TTSCode = "de-DE"
        TTSName = "de-DE-Studio-B"
    } else if language == "english" {
        TTSCode = "en-GB"
        TTSName = "en-GB-Studio-B"

    }
     // Build the request without effects profile
    req := &texttospeechpb.SynthesizeSpeechRequest{
        Input: &texttospeechpb.SynthesisInput{
            InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
        },
        Voice: &texttospeechpb.VoiceSelectionParams{
            // LanguageCode: "de-DE",
            // Name:         "de-DE-Studio-B",
			LanguageCode: 	TTSCode,
   			Name: 			TTSName,
        },
        AudioConfig: &texttospeechpb.AudioConfig{
            AudioEncoding: texttospeechpb.AudioEncoding_LINEAR16,
        },
    }


    // log.Printf("SynthesizeSpeechRequest: %+v\n", req)

    // Perform the text-to-speech request
    response, err := client.SynthesizeSpeech(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to synthesize speech: %w", err)
    }

    // log.Printf("SynthesizeSpeechResponse: %+v\n", response)

    // Write the response to a file
    outFile, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create output file: %w", err)
    }
    defer outFile.Close()
    _, err = outFile.Write(response.AudioContent)
    if err != nil {
        return fmt.Errorf("failed to write audio content to file: %w", err)
    }

    return nil
}
// ListVoices lists available voices for a given language code
func ListVoices(languageCode string) ([]*texttospeechpb.Voice, error) {
    ctx := context.Background()

    client, err := texttospeech.NewClient(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create TTS client: %w", err)
    }
    defer client.Close()

    req := &texttospeechpb.ListVoicesRequest{
        LanguageCode: languageCode,
    }

    resp, err := client.ListVoices(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to list voices: %w", err)
    }

    return resp.Voices, nil
}