# Kuro Mocks Web Interface

A modern, responsive web interface for managing Kuro mock services.

## Features

- **Dashboard**: View all mock services at a glance
- **Service Management**: Start/stop individual mock services
- **Real-time Status**: See which services are running
- **Responsive Design**: Works on desktop and mobile devices
- **PWA Support**: Install as a desktop/mobile app

## Getting Started

1. **Start the web interface**:
   ```bash
   usekuro web 8798
   ```

2. Open your browser to: [http://localhost:8798](http://localhost:8798)

## Development

The web interface is built with:

- [Vue.js 3](https://v3.vuejs.org/) - Progressive JavaScript framework
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [Font Awesome](https://fontawesome.com/) - Icons

### Project Structure

```
web/
├── static/
│   ├── css/
│   │   └── styles.css     # Custom styles
│   ├── js/
│   │   ├── app.js        # Main application logic
│   │   └── sw.js         # Service worker for PWA
│   └── site.webmanifest  # PWA manifest
└── index.html            # Main HTML file
```

## API Endpoints

- `GET /api/mocks` - List all mock services
- `POST /api/mocks` - Add a new mock service
- `POST /api/mocks/{id}/toggle` - Toggle a mock service
- `POST /api/server/toggle` - Toggle the mock server

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
