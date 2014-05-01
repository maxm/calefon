const int max_ranges = 30;
const int relay = D6;
const int button = A6;

unsigned long buttonStart = 0;
bool buttonPrevious = false;
unsigned long lastButtonPress = 0;

unsigned long buttonDuration = 60*60*1000;

int relayNow = LOW;

struct range {
  long start;
  long end;
};
range ranges[max_ranges];
unsigned long reference_time;

TCPClient client;
byte server[] = { 162, 243, 62, 133 };
int port = 9001;
unsigned long lastServerUpdate = 0;
unsigned long serverUpdateFrequency = 15*60*1000;

range new_ranges[max_ranges];
unsigned long new_reference_time;
unsigned int bufferPos = 0;

void setup() {
  pinMode(relay, OUTPUT);
  pinMode(button, INPUT_PULLUP);
  
  digitalWrite(relay, relayNow);
  
  for (int i = 0; i < max_ranges; ++i) {
      ranges[i].start = ranges[i].end = 0;
  }
}

void loop() {
  // Check button
  bool buttonNow = digitalRead(button) == LOW;
  if (!buttonPrevious && buttonNow) {
    buttonStart = millis();
  } else if (buttonPrevious && !buttonNow) {
    if (millis() - buttonStart > 20) {
        lastButtonPress = millis();
    }
  }
  buttonPrevious = buttonNow;

  // Update ranges from server
  if (client.connected()) {
    while (client.available() > 0 && bufferPos < sizeof(new_ranges)) {
      ((char*)(&new_ranges))[bufferPos] = client.read();
      ++bufferPos;
      if (bufferPos > 0 && bufferPos % sizeof(range) == 0) {
        int index = bufferPos / sizeof(range) - 1;
        if (new_ranges[index].start == 0 && new_ranges[index].end == 0) {
          client.stop();
          reference_time = new_reference_time;
          for (int i = 0; i < max_ranges; ++i) {
            if (i <= index) {
              ranges[i].start = new_ranges[i].start;
              ranges[i].end = new_ranges[i].end;
            } else {
              ranges[i].start = ranges[i].end = 0;
            }
          }
        }
      }
    }
  } else if (lastServerUpdate == 0 || millis() - lastServerUpdate > serverUpdateFrequency) {
    client.connect(server, port);
    new_reference_time = millis();
    bufferPos = 0;
    lastServerUpdate = millis();
  }
  
  // Update relay
  relayNow = LOW;
  for (int i = 0; i < max_ranges; ++i) {
    long t = millis() - reference_time;
    if (t >= ranges[i].start && t < ranges[i].end) {
      relayNow = HIGH;
    }
  }
  if (lastButtonPress != 0 && millis() - lastButtonPress < buttonDuration) {
    relayNow = HIGH;
  }
  digitalWrite(relay, relayNow);
}
