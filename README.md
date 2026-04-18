<div align="center">
  <a href="https://codeberg.org/pranaovs/Qashare">  </a>

<h1 align="center">Qashare</h1>

  <p align="center">
    A (mobile) app to track shared expenses and split bills easily.

[![Stargazers][stars-badge]][stars-url]
[![Stargazers][stars-github-badge]][stars-github-url]
[![Issues][issues-badge]][issues-url]
![Last Commit Badge][last-commit-badge]
[![AGPL-3.0 License][license-badge]][license-url]


  </p>
    <p align="center">
    <a href="https://codeberg.org/pranaovs/Qashare/issues">Report Bug</a>
    <a href="https://codeberg.org/pranaovs/Qashare/wiki">View Docs</a>
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

[![Go CI](https://codeberg.org/pranaovs/Qashare/badges/workflows/go-ci.yml/badge.svg?logo=go&label=Go%20CI&branch=main)](https://codeberg.org/pranaovs/Qashare/actions?workflow=go-ci.yml)

- Go >= 1.25.3 (test for lower)
- PostgreSQL database

#### Client (Flutter App)

[![Flutter CI](https://codeberg.org/pranaovs/Qashare/badges/workflows/flutter-ci.yml/badge.svg?logo=flutter&label=Flutter%20CI&branch=main)](https://codeberg.org/pranaovs/Qashare/actions?workflow=flutter-ci.yml)

- Flutter SDK >= 3.9.2 (test for lower)
- Dart SDK (included with Flutter)

### Installation

1. Clone the repo

   ```sh
   git clone https://codeberg.org/pranaovs/Qashare.git
   ```

2. Switch to the directory

    ```sh
    cd Qashare
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


[![Container Build badge](https://codeberg.org/pranaovs/Qashare/badges/workflows/container-build.yml/badge.svg?logo=docker&label=Container%20Build&branch=main)](https://codeberg.org/pranaovs/Qashare/actions?workflow=container-build.yml)

1. Download and edit the [`docker-compose.yml`](https://codeberg.org/pranaovs/Qashare/src/branch/main/server/docker-compose.yml) file to set your environment variables.

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

The Qashare API is documented using Swagger/OpenAPI. You can access the Swagger UI by navigating to:
<https://qashare.pranaovs.me/swagger/index.html>

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
- [x] Proper logout flow
- [x] Create frontend (not the vibe-coded slop)
- [x] Payment settlement
    - [x] Frontend integration
    - [x] Settlements in db
    - [x] Settlement verification (only payer can modify)
- [x] Bill splitting algorithms
- [ ] Group management features
- [x] User spending reports
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

This project follows [Semantic Commit Messages](https://gist.github.com/joshbuchea/6f47e86d2510bce28f8e7f42ae84c716) for the commits.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat(category): Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the AGPL-3.0 License. See [`LICENSE`][license-url] for more information.

## Maintainers

- Sasvat S R ([client/](https://codeberg.org/pranaovs/Qashare/src/branch/main/client)) - [@sasvat](https://codeberg.org/sasvat)
- Pranaov S ([server/](https://codeberg.org/pranaovs/Qashare/src/branch/main/server)) - [@pranaovs](https://codeberg.org/pranaovs)

## Contact

Pranaov S - [@pranaovs](mailto:qashare.contact@pranaovs.me)

Repo Link: [https://codeberg.org/pranaovs/Qashare](https://codeberg.org/pranaovs/Qashare)

## Acknowledgments

- [othneildrew (README Template)](https://github.com/othneildrew/Best-README-Template)


<!-- MARKDOWN LINKS & IMAGES -->
<!-- [stars-badge]: https://img.shields.io/gitea/stars/pranaovs/Qashare?gitea_url=https://codeberg.org&logo=codeberg -->
[stars-badge]: https://codeberg.org/pranaovs/Qashare/badges/stars.svg?logo=codeberg&style=social
[stars-url]: https://codeberg.org/pranaovs/Qashare/stars
[stars-github-badge]: https://img.shields.io/github/stars/pranaovs/Qashare
[stars-github-url]: https://github.com/pranaovs/Qashare/stargazers
[last-commit-badge]: https://img.shields.io/gitea/last-commit/pranaovs/Qashare?gitea_url=https://codeberg.org
[issues-badge]: https://codeberg.org/pranaovs/Qashare/badges/issues/open.svg
[issues-url]: https://codeberg.org/pranaovs/Qashare/issues
[license-badge]: https://img.shields.io/badge/License-AGPL_v3-red.svg
[license-url]: https://codeberg.org/pranaovs/Qashare/src/branch/main/LICENSE
[flutter-badge]: https://img.shields.io/badge/Flutter-027DFD?logo=flutter&logoColor=white
[flutter-url]: https://flutter.dev/
[go-badge]: https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white
[go-url]: https://go.dev/
[postgresql-badge]: https://img.shields.io/badge/PostgreSQL-316192?logo=postgresql&logoColor=white
[postgresql-url]: https://www.postgresql.org/
