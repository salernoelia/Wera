import requests
from pydub import AudioSegment
from pydub.playback import play
import io

def fetch_and_play_audio(url):
    # Send a GET request to get the audio file directly
    response = requests.get(url)
    if response.status_code == 200:
        # Load the audio file from binary data
        audio_data = io.BytesIO(response.content)
        song = AudioSegment.from_file(audio_data, format="wav")
        
        # Play the audio
        print("Playing audio...")
        play(song)
    else:
        print("Failed to fetch audio:", response.status_code)

# Replace with the actual URL or route you want to hit
fetch_and_play_audio("http://192.168.1.13:8080/weather")
