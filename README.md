# 🚀 EDL Telemetry Reconstructor

## Description
The EDL (Entry, Descent, and Landing) Telemetry Reconstructor is a robust, end-to-end simulation and data fusion pipeline designed to model and reconstruct spacecraft descent telemetry. This project generates noisy simulated sensor data (IMU and Radar) based on real physics, processes it through a high-throughput Extended Kalman Filter (EKF) written in Go, and visualizes the reconstructed trajectory and orientation in a real-time 3D React dashboard.

## 📑 Table of Contents
- [Features](#-features)
- [Technologies Used](#-technologies-used)
- [Installation](#-installation)
- [Usage](#-usage)
- [Project Structure](#-project-structure)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features
* **Physics & Sensor Simulation:** A Python-based flight simulator models aerodynamic drag, gravity, and thrust while generating realistic IMU (Gaussian noise, bias, random walk) and radar altimeter data.
* **High-Throughput Sensor Fusion:** A CPU-optimized Go pipeline consumes raw UDP/Kafka streams and uses an Extended Kalman Filter to fuse IMU and Radar data, producing clean state vectors (Quaternions + Position) at 60Hz.
* **3D Visualization Dashboard:** A React and TypeScript UI uses Three.js to render the spacecraft and Martian terrain in 3D, alongside an Altitude/Velocity chart comparing raw noise to the Kalman-filtered truth.
* **Mission Playback:** Includes a timeline scrubber allowing operators to rewind and analyze specific descent events seamlessly.
* **Scalable Infrastructure:** Built to handle massive telemetry throughput using Kafka for message brokering and TimescaleDB for storing raw and filtered datasets.

## 🛠️ Technologies Used
* **Simulation Engine:** Python
* **Telemetry Pipeline & EKF:** Go, Gonum (for linear algebra)
* **Frontend Dashboard:** React, TypeScript, Three.js (@react-three/fiber), Recharts
* **Data Infrastructure:** Kafka, TimescaleDB, WebSockets
* **Deployment:** Docker, Docker Compose, Kubernetes

## ⚙️ Installation

1. Clone the repository:
   ```bash
   git clone [https://github.com/yourusername/edl-telemetry-reconstructor.git](https://github.com/yourusername/edl-telemetry-reconstructor.git)
   cd edl-telemetry-reconstructor
   ```

2. Boot the local infrastructure (Simulator, Go Pipeline, React UI, Kafka, and Database) using Docker Compose:
   ```bash
   docker-compose up -d
   ```

3. Ensure all services are running:
   ```bash
   docker-compose ps
   ```

## 💻 Usage

* **Access the Dashboard:** Open your browser and navigate to `http://localhost:3000` (or your configured port) to access the React-based 3D visualization dashboard.
* **Review Telemetry:** Observe the artificial horizon/navball and the Altitude vs. Velocity charts as the Go pipeline reconstructs the noisy descent data in real-time.
* **Playback Mode:** Use the Timeline Scrubber in the UI to pause, rewind, and replay specific anomalies or phases of the EDL sequence.

## 📂 Project Structure
* `/flight_simulator`: Python application serving as the "Spacecraft," generating true descent physics and noisy sensor streams.
* `/telemetry_pipeline`: Go-based high-throughput ingestion engine executing the Extended Kalman Filter math and WebSocket broadcasting.
* `/mission_playback_ui`: React/TypeScript 3D visualization frontend featuring Three.js rendering and interactive charts.
* `/infrastructure`: YAML manifests for deploying the Kafka cluster, TimescaleDB, and fusion engine via Kubernetes.
* `/.github/workflows`: CI/CD pipelines including rigorous unit tests for the EKF matrix math.

## 🤝 Contributing
Contributions, issues, and feature requests are welcome! Feel free to check the issues page. If contributing code, please ensure that the Kalman filter matrix tests pass via the included GitHub Actions workflow.

## 📄 License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
