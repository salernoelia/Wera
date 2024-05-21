#include <WiFi.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include <Wire.h>
#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>

// Declarations (Prototypes) for all custom functions
void fetchData();
void manageWiFiConnection();
String replaceUmlauts(String input);

// Global variables and configuration constants
const char* ssid = "Epstein";
const char* password = "Passwort123";
const char* serverUrl = "https://spatial-interaction.onrender.com/cityclimate";
#define SCREEN_WIDTH 128
#define SCREEN_HEIGHT 64
#define OLED_RESET -1
#define OLED_I2C_ADDRESS 0x3D 
Adafruit_SSD1306 display(SCREEN_WIDTH, SCREEN_HEIGHT, &Wire, OLED_RESET);

const int cycleButtonPin = 18;  // Button to cycle through data points
const int refreshButtonPin = 23;  // Button to refresh current data point
int cycleButtonState;
int lastCycleButtonState = LOW;
int refreshButtonState;
int lastRefreshButtonState = LOW;
unsigned long lastDebounceTime = 0;
unsigned long debounceDelay = 50;
int currentIndex = 0;      // Start at index 0

void setup() {
  Serial.begin(115200);
  pinMode(cycleButtonPin, INPUT_PULLUP);
  pinMode(refreshButtonPin, INPUT_PULLUP);

  Serial.println("Initializing OLED display...");
  if (!display.begin(SSD1306_SWITCHCAPVCC, OLED_I2C_ADDRESS)) {
    Serial.println(F("SSD1306 allocation failed"));
    for (;;);
  }
  display.display();
  delay(2000);
  display.clearDisplay();
  Serial.println("OLED display initialized");

  manageWiFiConnection();

  
  Serial.println("Connected to WiFi");

  fetchData(); // Fetch initial data on setup
}

void loop() {
  int cycleReading = digitalRead(cycleButtonPin);
  int refreshReading = digitalRead(refreshButtonPin);

  if (cycleReading != lastCycleButtonState || refreshReading != lastRefreshButtonState) {
    lastDebounceTime = millis();
  }

  if ((millis() - lastDebounceTime) > debounceDelay) {
    if (cycleReading != cycleButtonState) {
      cycleButtonState = cycleReading;
      if (cycleButtonState == LOW) {
        currentIndex = (currentIndex + 1) % 61;
        fetchData();
      }
    }
    
    if (refreshReading != refreshButtonState) {
      refreshButtonState = refreshReading;
      if (refreshButtonState == LOW) {
        manageWiFiConnection();
      }
    }
  }

  lastCycleButtonState = cycleReading;
  lastRefreshButtonState = refreshReading;
}

String replaceUmlauts(String input) {
  input.replace("ü", "ue");
  input.replace("ä", "ae");
  input.replace("ö", "oe");
  input.replace("Ü", "Ue");
  input.replace("Ä", "Ae");
  input.replace("Ö", "Oe");
  input.replace("ß", "ss");
  return input;
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
    if (currentIndex < features.size()) {
      JsonObject feature = features[currentIndex];
      const char* name = feature["properties"]["name"];
      double value = feature["properties"]["values"];

      String displayName = replaceUmlauts(String(name));

      display.clearDisplay();
      display.setTextSize(1);
      display.setCursor(0, 0);
      display.println(displayName);
      display.setTextSize(2);
      display.setCursor(0, 20);
      display.print(value, 1);
      display.print("C");
      display.display();
    } else {
      display.println("Index out of range");
      display.display();
    }
  } else {
    display.println("Fetch failed!");
    display.println("Error code: " + String(httpResponseCode));
    display.display();
  }

  http.end();
}

void manageWiFiConnection() {


  if (WiFi.status() == WL_CONNECTED) {
    display.clearDisplay();
  display.setTextSize(1);
  display.setTextColor(SSD1306_WHITE);
  display.setCursor(0, 0);
  display.println("Wifi is Connected!");
  display.display();

    return;
  }


  Serial.println("Disconnecting WiFi...");
    WiFi.disconnect();


  WiFi.begin(ssid, password);


  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    display.println("Connecting to WiFi...");

  }
  display.clearDisplay();
  display.setCursor(0, 0);
  display.display();
}
