/// URL base de la API. En web/desktop apunta a localhost; en un emulador
/// Android real usar 10.0.2.2 en vez de localhost (loopback del host).
const apiBaseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'http://localhost:8080');
