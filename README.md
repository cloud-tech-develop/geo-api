# Geo API

API REST en Go para consultar países, departamentos/estados y ciudades del mundo.

## Stack

- **Go 1.22** — net/http estándar (sin frameworks externos)
- **Dataset** — [countries-states-cities-database](https://github.com/dr5hn/countries-states-cities-database) (~150k ciudades)
- **Dependencias externas**: ninguna

---

## Instalación

### 1. Clonar y descargar el dataset

```bash
git clone https://github.com/youruser/geo-api
cd geo-api

# Crear carpeta de datos
mkdir -p data

# Descargar el JSON del dataset (requiere curl)
curl -L "https://raw.githubusercontent.com/dr5hn/countries-states-cities-database/master/json/countries%2Bstates%2Bcities.json" -o data/countries+states+cities.json
```

### 2. Levantar el servidor

```bash
go run ./cmd/api
```

El servidor inicia en `http://localhost:8082`.

### 3. Con Docker

```bash
docker build -t geo-api .
docker run -p 8082:8082 geo-api
```

### 4. Variables de entorno

| Variable        | Default                             | Descripción              |
| --------------- | ----------------------------------- | ------------------------ |
| `PORT`          | `8082`                              | Puerto del servidor      |
| `GEO_DATA_PATH` | `data/countries+states+cities.json` | Ruta al archivo de datos |

---

## Endpoints

### Salud y estadísticas

| Método | Ruta    | Descripción                               |
| ------ | ------- | ----------------------------------------- |
| GET    | /health | Estado del servidor                       |
| GET    | /stats  | Conteo total de países, estados, ciudades |

### Países

| Método | Ruta                     | Descripción                        |
| ------ | ------------------------ | ---------------------------------- |
| GET    | /countries               | Lista todos los países             |
| GET    | /countries/{iso2}        | Detalle de un país por código ISO2 |
| GET    | /countries/{iso2}/states | Estados/departamentos del país     |
| GET    | /countries/{iso2}/cities | Todas las ciudades del país        |

### Estados / Departamentos

| Método | Ruta                | Descripción                 |
| ------ | ------------------- | --------------------------- |
| GET    | /states/{id}        | Detalle de un estado por ID |
| GET    | /states/{id}/cities | Ciudades del estado         |

### Query params comunes

| Param    | Tipo   | Default | Descripción                           |
| -------- | ------ | ------- | ------------------------------------- |
| `search` | string | —       | Filtrar por nombre (case-insensitive) |
| `page`   | int    | 1       | Página                                |
| `limit`  | int    | 50      | Resultados por página (máx 500)       |

---

## Ejemplos de uso

```bash
# Todos los países
curl http://localhost:8082/countries

# Buscar países con "col" en el nombre
curl "http://localhost:8082/countries?search=col"

# Detalle de Colombia
curl http://localhost:8082/countries/CO

# Departamentos de Colombia
curl http://localhost:8082/countries/CO/states

# Buscar departamentos con "antioquia"
curl "http://localhost:8082/countries/CO/states?search=antioquia"

# Ciudades del estado con ID 2877 (Antioquia)
curl http://localhost:8082/states/2877/cities

# Buscar ciudades con "medellin" en Colombia
curl "http://localhost:8082/countries/CO/cities?search=medellin"

# Estadísticas del dataset
curl http://localhost:8082/stats
```

---

## Estructura del proyecto

```
geo-api/
├── cmd/
│   └── api/
│       └── main.go          # Entry point, servidor HTTP
├── internal/
│   ├── handler/
│   │   └── handler.go       # Handlers HTTP y router
│   ├── model/
│   │   └── geo.go           # Structs de datos
│   └── repository/
│       └── geo.go           # Carga de datos e índices en memoria
├── data/
│   └── countries+states+cities.json   # Dataset (descargar manualmente)
├── Dockerfile
├── go.mod
└── README.md
```

---

## Respuesta de ejemplo

### `GET /countries/CO`

```json
{
  "id": 48,
  "name": "Colombia",
  "iso2": "CO",
  "iso3": "COL",
  "phone_code": "57",
  "capital": "Bogotá",
  "currency": "COP",
  "region": "Americas",
  "subregion": "South America",
  "emoji": "🇨🇴"
}
```

### `GET /countries/CO/states` (paginado)

```json
{
  "data": [
    { "id": 2877, "name": "Antioquia", "state_code": "ANT", "country_code": "CO", ... },
    ...
  ],
  "total": 33,
  "page": 1,
  "limit": 50
}
```
