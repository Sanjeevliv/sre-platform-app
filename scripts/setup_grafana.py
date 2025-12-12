import requests
import json
import glob
import os
import time

GRAFANA_URL = "http://localhost:3000"
AUTH = ("admin", "admin")
HEADERS = {"Content-Type": "application/json"}

def wait_for_grafana():
    print("Waiting for Grafana to be ready...")
    for _ in range(30):
        try:
            r = requests.get(f"{GRAFANA_URL}/api/health", timeout=2)
            if r.status_code == 200:
                print("Grafana is up!")
                return True
        except requests.exceptions.ConnectionError:
            pass
        time.sleep(2)
    return False

def create_datasource():
    print("Creating Prometheus Datasource...")
    payload = {
        "name": "Prometheus",
        "type": "prometheus",
        "url": "http://sre-prometheus:9090",
        "access": "proxy",
        "isDefault": True
    }
    r = requests.post(f"{GRAFANA_URL}/api/datasources", auth=AUTH, headers=HEADERS, json=payload)
    if r.status_code in [200, 409]: # 200 created, 409 exists
        print("Datasource created or already exists.")
    else:
        print(f"Failed to create datasource: {r.text}")

def import_dashboards():
    files = glob.glob("charts/sre-platform/dashboards/*.json")
    for f_path in files:
        print(f"Importing {f_path}...")
        with open(f_path, 'r') as f:
            dashboard_json = json.load(f)
        
        # Nullify ID to allow creation/overwrite proper
        dashboard_json["id"] = None
        
        payload = {
            "dashboard": dashboard_json,
            "overwrite": True
        }
        
        r = requests.post(f"{GRAFANA_URL}/api/dashboards/db", auth=AUTH, headers=HEADERS, json=payload)
        if r.status_code == 200:
            print(f"Success! URL: {GRAFANA_URL}{r.json().get('url')}")
        else:
            print(f"Failed to import {f_path}: {r.text}")

if __name__ == "__main__":
    if wait_for_grafana():
        create_datasource()
        import_dashboards()
        print("\nSETUP COMPLETE. Login with admin/admin.")
    else:
        print("Could not connect to Grafana.")
