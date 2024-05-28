#include <HardwareSerial.h>

HardwareSerial GPS(1); // Use the second hardware serial port on ESP32

void setup() {
  Serial.begin(115200);
  GPS.begin(9600, SERIAL_8N1, 16, 17); // Initialize GPS on RX pin 16 and TX pin 17
}

String gpsDataBlock = "";

void loop() {
  while (GPS.available()) {
    char c = GPS.read();
    if (c == '\n') {
      if (gpsDataBlock.startsWith("$GNRMC") || gpsDataBlock.startsWith("$GNGLL")) {
        processGPSData(gpsDataBlock);
      }
      gpsDataBlock = ""; // Reset the data block after processing
    } else {
      gpsDataBlock += c; // Accumulate data until newline
    }
  }
  delay(100);
}

void processGPSData(String data) {
  if (data.startsWith("$GNRMC")) {
    int status = data.indexOf('A', 7); // Look for 'A' indicating valid data
    if (status > 0) {
      String time = getValue(data, ',', 1);
      String latitude = getValue(data, ',', 3);
      String ns = getValue(data, ',', 4);
      String longitude = getValue(data, ',', 5);
      String ew = getValue(data, ',', 6);

      float lat = convertToDecimalDegrees(latitude, true);
      float lon = convertToDecimalDegrees(longitude, false);
      if (ns == "S") lat = -lat;
      if (ew == "W") lon = -lon;

      // Check if coordinates are zero
      if (lat == 0.0 && lon == 0.0) {
        Serial.println("Invalid GPS Data: Skipping zero coordinates.");
        return; // Skip processing this data
      }

      Serial.print("Valid GPS Data: ");
      Serial.print("Time: " + time);
      Serial.print(", Latitude: " + String(lat, 6));
      Serial.print(", Longitude: " + String(lon, 6));
      Serial.println();
    }
  }
}

String getValue(String data, char separator, int index) {
  int found = 0;
  int strIndex[] = {0, -1};
  int maxIndex = data.length() - 1;

  for (int i = 0; i <= maxIndex && found <= index; i++) {
    if (data.charAt(i) == separator || i == maxIndex) {
      found++;
      strIndex[0] = strIndex[1] + 1;
      strIndex[1] = (i == maxIndex) ? i+1 : i;
    }
  }
  return found > index ? data.substring(strIndex[0], strIndex[1]) : "";
}

float convertToDecimalDegrees(String coordinate, bool isLatitude) {
  int degreeLength = isLatitude ? 2 : 3;
  float degrees = coordinate.substring(0, degreeLength).toFloat();
  float minutes = coordinate.substring(degreeLength).toFloat() / 60.0;
  return degrees + minutes;
}