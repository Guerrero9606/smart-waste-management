import os
import time
import random
import requests
from datetime import datetime, timezone
from dotenv import load_dotenv

def main():
    """
    Función principal que ejecuta el simulador de sensores IoT.
    """
    print("--- Iniciando Simulador de Sensores de Contenedores ---")

    dotenv_path = os.path.join(os.path.dirname(__file__), '..', '.env')
    load_dotenv(dotenv_path=dotenv_path)

    api_base_url = os.getenv("API_BASE_URL", "http://localhost:8080")
    api_endpoint = f"{api_base_url}/api/v1/readings"
    
    try:
        interval_seconds = int(os.getenv("SIMULATOR_INTERVAL_SECONDS", 10))
    except (ValueError, TypeError):
        print("ADVERTENCIA: SIMULATOR_INTERVAL_SECONDS no es un número válido. Usando valor por defecto de 10 segundos.")
        interval_seconds = 10

    container_ids = [
        "c7a1c7d6-3d2c-4e8d-9a6a-0b1e4c7b8e1a",
        "b8b2d8e7-4e3d-5f9e-a0b0-1c2f5d8e9f2b",
        "a9c3e9f8-5f4e-6a0f-b1c1-2d3a6e9f0a3c",
    ]

    print(f"URL del API de Ingesta: {api_endpoint}")
    print(f"Número de contenedores a simular: {len(container_ids)}")
    print(f"Intervalo entre ciclos de envío: {interval_seconds} segundos")
    print("---------------------------------------------------------")

    while True:
        print(f"\nIniciando nuevo ciclo de envío de lecturas a las {datetime.now().strftime('%H:%M:%S')}...")
        
        for container_id in container_ids:
            send_reading(api_endpoint, container_id)
            time.sleep(random.uniform(0.2, 1.0))

        print(f"--- Ciclo completado. Esperando {interval_seconds} segundos para el siguiente. ---")
        time.sleep(interval_seconds)


def send_reading(url: str, container_id: str):
    """
    Genera una lectura de llenado aleatoria y la envía al endpoint de la API.
    """
    try:
        fill_level = int(random.triangular(5, 100, 70))
        
        timestamp = datetime.now(timezone.utc).isoformat()

        payload = {
            "container_id": container_id,
            "fill_level": fill_level,
            "timestamp": timestamp
        }

        response = requests.post(url, json=payload, timeout=5)

        if response.status_code == 202:
            print(f"  [OK] Contenedor {container_id[:8]}: Lectura enviada (Nivel: {fill_level}%)")
        else:
            print(f"  [ERROR] Contenedor {container_id[:8]}: Respuesta inesperada. "
                  f"Status: {response.status_code}, Body: {response.text}")

    except requests.exceptions.RequestException as e:
        print(f"  [FATAL] Contenedor {container_id[:8]}: No se pudo conectar a la API. Error: {e}")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n--- Simulador detenido por el usuario. ¡Adiós! ---")
        exit(0)