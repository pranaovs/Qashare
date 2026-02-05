<div align="center">
  <a href="https://github.com/pranaovs/qashare">  </a>

<h1 align="center">Qashare</h1>

  <p align="center">
    A (mobile) app to track shared expenses and split bills easily.

[![Stargazers][stars-badge]][stars-url]
[![Forks][forks-badge]][forks-url]
[![Discussions][discussions-badge]][discussions-url]
[![Issues][issues-badge]][issues-url]
![Last Commit Badge][last-commit-badge]
[![AGPL-3.0 License][license-badge]][license-url]

  </p>
    <p align="center">
    <a href="https://github.com/pranaovs/qashare"></a>
    <a href="https://github.com/pranaovs/qashare/issues">Report Bug</a>
    <a href="https://github.com/pranaovs/qashare/wiki">View Docs</a>
  </p>
</div>

<!--toc:start-->
- [About The Project](#about-the-project)
  - [Built Using](#built-using)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
    - [Server (Backend)](#server-backend)
    - [Client (Flutter App)](#client-flutter-app)
  - [Installation](#installation)
    - [Running the Server](#running-the-server)
    - [Running the Server with Docker](#running-the-server-with-docker)
    - [Installing the client](#installing-the-client)
- [API Documentation](#api-documentation)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Maintainers](#maintainers)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)
<!--toc:end-->

## About The Project

> Pronounced: Cash-Share

A full-stack application designed to help users track shared expenses and split bills easily. The app features a Flutter mobile client and a Go backend with PostgreSQL database.

### Built Using

- [![Flutter][flutter-badge]][flutter-url]
- [![Go][go-badge]][go-url]
- [![PostgreSQL][postgresql-badge]][postgresql-url]

## Getting Started

To set up the project locally, follow these instructions for both the client and server components.

### Prerequisites

#### Server (Backend)

[![Container Build](https://github.com/pranaovs/Qashare/actions/workflows/container-release.yml/badge.svg)](https://github.com/pranaovs/Qashare/actions/workflows/container-release.yml)

- Go >= 1.25.3 (test for lower)
- PostgreSQL database

#### Client (Flutter App)

[![Flutter Build](https://github.com/pranaovs/Qashare/actions/workflows/flutter-build.yml/badge.svg)](https://github.com/pranaovs/Qashare/actions/workflows/flutter-build.yml)

- Flutter SDK >= 3.9.2 (test for lower)
- Dart SDK (included with Flutter)

### Installation

1. Clone the repo

   ```sh
   git clone https://github.com/pranaovs/qashare.git
   ```

2. Switch to the directory

    ```sh
    cd qashare
    ```

#### Running the Server

1. Switch to the project directory

    ```sh
    cd server
    ```

2. Install the dependencies

    ```sh
    go get
    ```

3. Run the app

    ```sh
    go run .
    ```

#### Running the Server with Docker

1. Download and edit the [`docker-compose.yml`](https://github.com/pranaovs/Qashare/blob/main/server/docker-compose.yml) file to set your environment variables.

2. Run Docker Compose

    ```sh
    docker compose up
    ```

#### Installing the client

1. Switch to the project directory

    ```sh
    cd client
    ```

2. Install the dependencies

    ```sh
    flutter pub get
    ```

3. Run the app

    ```sh
    flutter run
    ```

## API Documentation

The Qashare API is documented using Swagger/OpenAPI. Once the server is running, you can access the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

The Swagger UI provides:
- Complete API endpoint documentation
- Request/response schemas for all operations
- Interactive testing capability (try out API calls directly from the browser)
- Authentication using JWT Bearer tokens

## Roadmap

- [x] Set up Flutter client structure
- [x] Set up Go backend with Gin framework
- [x] Implement PostgreSQL database integration
- [x] User authentication and authorization
- [x] Expense tracking and management
- [ ] Proper logout flow
- [x] Create frontend (not the vibe-coded slop)
- [x] Payment settlement
    - [ ] Frontend integration
    - [ ] Settlements in db
    - [ ] Settlement verification
- [x] Bill splitting algorithms
- [ ] Group management features
- [ ] User spending reports
- [x] Guest Users
- [ ] Permission management
- [ ] Edit history
- [ ] Data import/export
- [ ] Statements generation
- [ ] Bundle server in client for fully-local usage
  - [ ] Upload to cloud option
  - [ ] Open embedded server to LAN

<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request.
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the AGPL-3.0 License. See [`LICENSE`](https://github.com/pranaovs/qashare/blob/main/LICENSE) for more information.

## Maintainers

- Sasvat S R ([client/](https://github.com/pranaovs/Qashare/tree/main/client)) - [@sasvat007](https://github.com/sasvat007)
- Pranaov S ([server/](https://github.com/pranaovs/Qashare/tree/main/server)) - [@pranaovs](https://github.com/pranaovs)

## Contact

Pranaov S - [@pranaovs](mailto:contact.anoinihooqaq@pranaovs.me)

Repo Link: [https://github.com/pranaovs/qashare](https://github.com/pranaovs/qashare)

## Acknowledgments

- [othneildrew (README Template)](https://github.com/othneildrew/Best-README-Template)

<!-- MARKDOWN LINKS & IMAGES -->
[forks-badge]: https://img.shields.io/github/forks/pranaovs/qashare
[forks-url]: https://github.com/pranaovs/qashare/network/members
[stars-badge]: https://img.shields.io/github/stars/pranaovs/qashare
[stars-url]: https://github.com/pranaovs/qashare/stargazers
[last-commit-badge]: https://img.shields.io/github/last-commit/pranaovs/qashare/main
[issues-badge]: https://img.shields.io/github/issues/pranaovs/qashare
[issues-url]: https://github.com/pranaovs/qashare/issues
[discussions-badge]: https://img.shields.io/github/discussions/pranaovs/qashare
[discussions-url]: https://github.com/pranaovs/qashare/discussions
[license-badge]: https://img.shields.io/github/license/pranaovs/qashare
[license-url]: https://github.com/pranaovs/qashare/blob/main/LICENSE
[flutter-badge]: https://img.shields.io/badge/Flutter-027DFD?logo=flutter&logoColor=white
[flutter-url]: https://flutter.dev/
[go-badge]: https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white
[go-url]: https://go.dev/
[postgresql-badge]: https://img.shields.io/badge/PostgreSQL-316192?logo=postgresql&logoColor=white
[postgresql-url]: https://www.postgresql.org/
