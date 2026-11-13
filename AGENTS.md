# geofence-service

Consumidor Go que detecta entradas e saídas de cercas eletrônicas (geofences). Lê cada posição da fila `geofence-service`, que o RabbitMQ alimenta a partir do exchange `asset-find-service` (publicado pelo `marine-br/asset-find-service`). Para cada cerca da empresa da posição, decide se o veículo está dentro ou fora e grava um registro `IN` ou `OUT` na coleção `geofencehistories` quando o estado muda. O relatório de cercas da API NestJS (`marine-br/api`, `src/modules/report/geofence-report/`) lê essa coleção.

O serviço não publica mensagens. Todo o fluxo está em `main.go`; `utils/eventValidators/eventValidator.go` decide se houve mudança de estado.

## Dados compartilhados

- Lê `geofences` e grava `geofencehistories` no MongoDB `tracking`, o mesmo banco da API e dos outros serviços. As cercas são criadas pela API; aqui só se leem as que têm a empresa da posição em `companies` e `deleted` diferente de `true`.
- As structs `TrackerMsgType` e `GeofenceType` vêm de `marine-br/golib-utils` (`models/`). Mudar um campo lá muda o que este serviço decodifica.
- No Redis, a última transição de cada par veículo/cerca fica na chave `<vehicleId>:<geofenceId>`, sem prefixo, com expiração de 5 dias.

## Armadilhas

- **Ordem por `GPS_TIME`.** O estado é decidido pelo horário do GPS, não pela ordem de chegada. O serviço compara com o último registro antes e o primeiro depois daquele `GPS_TIME`; uma posição atrasada pode apagar um registro posterior que ficou redundante. Mudanças nessa lógica precisam continuar aceitando posições fora de ordem.
- **Ack em todo caso.** Toda mensagem recebe ack, inclusive quando o JSON é inválido ou a busca de cercas falha. Não há requeue nem dead-letter, então um erro descarta a posição.
- **Cerca malformada derruba o pod.** `utils/PolygonFromPrimitiveD/PoligonFromPrimitiveD.go` chama `log.Fatalf` quando o `geojson` não tem `features` ou tem coordenada de tipo inesperado. Só o primeiro `Polygon` de cada cerca é usado.
- **Coordenadas invertidas de propósito.** No GeoJSON a ordem é `[longitude, latitude]`; o código monta `Point{X: latitude, Y: longitude}` dos dois lados (cerca e posição). Mantenha os dois lados na mesma convenção.
- **Sem filtro de tipo de evento.** Qualquer mensagem com latitude e longitude diferentes de zero é avaliada, inclusive heartbeats. O asset-find-service manda heartbeats para outro exchange, e é isso que os mantém fora desta fila.
- **Variável obrigatória ausente encerra o processo com código 0** (`setups/env.setup.go`), então o pod reinicia sem parecer falha.
- O caminho do módulo é `github.com/marine-br/geoafence-service` (com o erro de grafia). Use esse caminho nos imports.

## Verificação e entrega

- `go test ./...` e `go build ./...`. Os workflows não rodam testes.
- **Merge em `main` faz deploy em produção** (`build-and-deploy-prod.yaml`: imagem `geofence-service:stable`, `helm upgrade` no namespace `gateway-prod`, e todo pod reinicia). Push em `staging` faz deploy em desenvolvimento (namespace `gateway-dev`).
- O chart, as variáveis de ambiente e o número de réplicas ficam em `marine-br/infrastructure`, em `helm-packages/apps/geofence-service-prod/` e `helm-packages/apps/geofence-service/`. O workflow usa o branch padrão do infrastructure no momento do deploy, então uma mudança lá só chega ao cluster no próximo deploy deste serviço.
