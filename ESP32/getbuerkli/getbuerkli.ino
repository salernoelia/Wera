#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <Wire.h>
#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>

// WiFi credentials
const char* ssid = "Epstein";
const char* password = "Passwort123";

// Server URL
const char* serverUrl = "http://192.168.135.136:8080/cityclimate";

// OLED display settings
#define SCREEN_WIDTH 128
#define SCREEN_HEIGHT 64
#define OLED_RESET -1
#define OLED_I2C_ADDRESS 0x3D 
Adafruit_SSD1306 display(SCREEN_WIDTH, SCREEN_HEIGHT, &Wire, OLED_RESET);

// Button setup
const int buttonPin = 23;
int buttonState = 0;
int lastButtonState = LOW;
unsigned long lastDebounceTime = 0;
unsigned long debounceDelay = 50;

void setup() {
  Serial.begin(115200);
  pinMode(buttonPin, INPUT_PULLUP); // Ensuring button is pulled up

  // Initialize OLED display
  Serial.println("Initializing OLED display...");
  if(!display.begin(SSD1306_SWITCHCAPVCC, OLED_I2C_ADDRESS)) {
    Serial.println(F("SSD1306 allocation failed"));
    for(;;);
  }
  display.display();
  delay(2000);
  display.clearDisplay();
  Serial.println("OLED display initialized");

  // Connect to Wi-Fi
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.println("Connecting to WiFi..");
  }
  Serial.println("Connected to WiFi");
}

void loop() {
  int reading = digitalRead(buttonPin);

  if (reading != lastButtonState) {
    lastDebounceTime = millis();  // reset debounce timer
  }

  if ((millis() - lastDebounceTime) > debounceDelay) {
    if (reading != buttonState) {
      buttonState = reading;

      if (buttonState == LOW) {  // Assuming button is connected to GND
        fetchData();
      }
    }
  }

  lastButtonState = reading;
}

void fetchData() {
  display.clearDisplay();
  display.setTextSize(1);
  display.setTextColor(SSD1306_WHITE);
  display.setCursor(0, 0);
  display.println("Fetching Data...");
  display.display();

  HTTPClient http;
  http.begin(serverUrl);
  int httpResponseCode = http.GET();

  if (httpResponseCode > 0) {
    String payload = http.getString();
    DynamicJsonDocument doc(1024);
    deserializeJson(doc, payload);

    JsonArray features = doc["features"];
    for (JsonObject feature : features) {
      const char* name = feature["properties"]["name"];
      if (strcmp(name, "Bürkliplatz") == 0) {
        double value = feature["properties"]["values"];
        display.clearDisplay();
        display.setTextSize(2);
        display.setCursor(0, 0);
        display.print("Buerkli: ");
        display.setCursor(0, 20);
        display.print(value, 1); // display 1 decimal place
        display.print("C");
        display.display();
        break;
      }
    }
  } else {
    display.println("Fetch failed!");
    display.println("Error code: " + String(httpResponseCode));
    display.display();
  }

  http.end();
}
