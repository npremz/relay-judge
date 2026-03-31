# Relay Judge Standalone

Repo autonome du juge Code Relay, sans l'application web.

## Contenu

- binaire Go pour evaluer des soumissions Python
- sujets JSON charges dynamiquement depuis `subjects/`
- exemples Python pour verifier rapidement le setup

## Prerequis

- Go 1.22+
- Python 3 disponible dans le `PATH`

## Demarrage

Lister les sujets:

```bash
go run ./cmd/relay-judge list
```

Executer un exemple:

```bash
go run ./cmd/relay-judge run --subject two-sum --workspace ./examples
```

Build local:

```bash
./build.sh
```

## Distribution

Le script de build genere:

- `dist/relay-judge` ou `dist/relay-judge.exe`
- `dist/subjects`
- `dist/relay-judge-<goos>-<goarch>.tar.gz`

Exemples:

```bash
GOOS=linux GOARCH=amd64 ./build.sh
GOOS=darwin GOARCH=arm64 ./build.sh
GOOS=windows GOARCH=amd64 ./build.sh
```

Pour deployer sur des postes de jeu Unix, prefere l'archive `.tar.gz`.

Le poste cible doit avoir:

- le binaire
- le dossier `subjects` a cote du binaire
- `python3` dans le `PATH`

Si besoin:

```bash
chmod +x relay-judge
./relay-judge list
```

## Organisation

- `cmd/relay-judge`: point d'entree CLI
- `internal/subject`: chargement et resolution des sujets
- `internal/engine`: orchestration Go + wrapper Python embarque
- `internal/checker`: verification des resultats
- `internal/scoring`: suggestion de scoring jury
- `subjects/`: sujets et tests
- `examples/`: soumissions d'exemple

Le detail du cadrage jury reste dans [`JURY_EVAL_PLAN.md`](./JURY_EVAL_PLAN.md).
