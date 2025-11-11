<div align="center">
  <a href="https://github.com/pranaovs/shared-expenses-app">  </a>

<h1 align="center">Shared Expenses App (Name Tentative)</h1>

  <p align="center">
    A mobile app to track shared expenses and split bills easily.

[![Stargazers][stars-badge]][stars-url]
[![Forks][forks-badge]][forks-url]
[![Discussions][discussions-badge]][discussions-url]
[![Issues][issues-badge]][issues-url]
![Last Commit Badge][last-commit-badge]
[![AGPL-3.0 License][license-badge]][license-url]

  </p>
    <p align="center">
    <a href="https://github.com/pranaovs/shared-expenses-app"></a>
    <a href="https://github.com/pranaovs/shared-expenses-app/issues">Report Bug</a>
    <a href="https://github.com/pranaovs/shared-expenses-app/wiki">View Docs</a>
  </p>
</div>

<!--toc:start-->
- [About The Project](#about-the-project)
  - [Inspiration](#inspiration)
  - [Built Using](#built-using)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
    - [Server (Backend)](#server-backend)
    - [Client (Flutter App)](#client-flutter-app)
  - [Installation](#installation)
    - [Installing the client](#installing-the-client)
    - [Running the Server](#running-the-server)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contributors](#contributors)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)
<!--toc:end-->

## About The Project

A full-stack application designed to help users track shared expenses and split bills easily. The app features a Flutter mobile client and a Go backend with PostgreSQL database.

### Inspiration

The other day, my friends and I went to a water park. Once the trip was over, we wanted to settle our expenses.
While walking on the street, my friend installed the \[wise-expense-splitting-app\] and started adding expenses. But it was riddled with ads. The worst part, we couldn't add more than 5 expenses without some limitation (I think it was an ad-break, or a delay asking us to upgrade).
I also didn't want my expenses to be stored on some third-party server for who knows when and for what other purpose.

That's where the idea of building a simple cross-platform self-hosted app came to me.
In a few weeks, I had to prepare a "project" for my Database Management Course. I decided to use that as an excuse to kickstart this idea.

### Built Using

- [![Flutter][flutter-badge]][flutter-url]
- [![Go][go-badge]][go-url]
- [![PostgreSQL][postgresql-badge]][postgresql-url]

## Getting Started

To set up the project locally, follow these instructions for both the client and server components.

### Prerequisites

#### Server (Backend)

- Go >= 1.25.3 (test for lower)
- PostgreSQL database

#### Client (Flutter App)

- Flutter SDK >= 3.9.2 (test for lower)
- Dart SDK (included with Flutter)

### Installation

1. Clone the repo

   ```sh
   git clone https://github.com/pranaovs/shared-expenses-app.git
   ```

2. Switch to the directory

    ```sh
    cd shared-expenses-app
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

## Roadmap

- [x] Set up Flutter client structure
- [x] Set up Go backend with Gin framework
- [x] Implement PostgreSQL database integration
- [x] User authentication and authorization
- [x] Expense tracking and management
- [ ] Proper logout flow
- [ ] Create frontend (not the vibe-coded slop)
- [ ] Payment settlement
- [ ] Bill splitting algorithms
- [ ] Group management features
- [ ] User spending reports
- [ ] Guest Users
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

Distributed under the AGPL-3.0 License. See [`LICENSE`](https://github.com/pranaovs/shared-expenses-app/blob/main/LICENSE) for more information.

## Contributors

Sasvat S R - [@sasvat007](https://github.com/sasvat007)
S S Kavinthra - [@kav1nthra](https://github.com/kav1nthra)

## Contact

Pranaov S - [@pranaovs](mailto://contact.anoinihooqaq@pranaovs.me)

Repo Link: [https://github.com/pranaovs/shared-expenses-app](https://github.com/pranaovs/shared-expenses-app)

## Acknowledgments

- [othneildrew (README Template)](https://github.com/othneildrew/Best-README-Template)

<!-- MARKDOWN LINKS & IMAGES -->
[forks-badge]: https://img.shields.io/github/forks/pranaovs/shared-expenses-app
[forks-url]: https://github.com/pranaovs/shared-expenses-app/network/members
[stars-badge]: https://img.shields.io/github/stars/pranaovs/shared-expenses-app
[stars-url]: https://github.com/pranaovs/shared-expenses-app/stargazers
[last-commit-badge]: https://img.shields.io/github/last-commit/pranaovs/shared-expenses-app/main
[issues-badge]: https://img.shields.io/github/issues/pranaovs/shared-expenses-app
[issues-url]: https://github.com/pranaovs/shared-expenses-app/issues
[discussions-badge]: https://img.shields.io/github/discussions/pranaovs/shared-expenses-app
[discussions-url]: https://github.com/pranaovs/shared-expenses-app/discussions
[license-badge]: https://img.shields.io/github/license/pranaovs/shared-expenses-app
[license-url]: https://github.com/pranaovs/shared-expenses-app/blob/main/LICENSE
[flutter-badge]: https://img.shields.io/badge/Flutter-027DFD?logo=flutter&logoColor=white
[flutter-url]: https://flutter.dev/
[go-badge]: https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white
[go-url]: https://go.dev/
[postgresql-badge]: https://img.shields.io/badge/PostgreSQL-316192?logo=postgresql&logoColor=white
[postgresql-url]: https://www.postgresql.org/
