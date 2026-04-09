# Relay Judge Standalone

Repo autonome du juge Code Relay, sans l'application web.

## Contenu

- binaire Go pour evaluer des soumissions Python et C
- sujets JSON charges dynamiquement depuis `subjects/`
- exemples Python pour verifier rapidement le setup

## Prerequis

- Go 1.22+
- Python 3 disponible dans le `PATH`
- un compilateur C (`cc`, `clang` ou `gcc`) disponible dans le `PATH` pour les sujets C

## Demarrage

Quelques commandes utiles:

```bash
go run ./cmd/relay-judge list
go run ./cmd/relay-judge run --subject two-sum --workspace ./examples
go run ./cmd/relay-judge ./examples/two_sum.py
go run ./cmd/relay-judge --stress ./examples/two_sum.py
go run ./cmd/relay-judge ./examples/sort_the_stack.c --cc clang
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
- `python3` dans le `PATH` pour les sujets Python
- `cc`/`clang`/`gcc` dans le `PATH` pour les sujets C

Si besoin:

```bash
chmod +x relay-judge
./relay-judge list
```

## Organisation

- `cmd/relay-judge`: point d'entree CLI
- `internal/subject`: chargement et resolution des sujets
- `internal/engine`: orchestration Go + runners Python/C
- `internal/checker`: verification des resultats
- `internal/scoring`: suggestion de scoring jury
- `scripts/generate_subjects.py`: generation des `subject.json`, notamment des gros cas `perf`
- `subjects/`: sujets et tests
- `examples/`: soumissions d'exemple

Le detail du cadrage jury reste dans [`JURY_EVAL_PLAN.md`](./JURY_EVAL_PLAN.md).

## Utilisation CLI

Le binaire supporte 4 facons principales d'etre lance:

1. Sans argument sur un terminal interactif: ouvre le mode interactif.
2. `relay-judge list`: liste les sujets disponibles.
3. `relay-judge run ...`: mode explicite avec flags.
4. `relay-judge <file.py|file.c>`: deduction automatique du sujet a partir du nom du fichier.

Exemples:

```bash
relay-judge
relay-judge list
relay-judge run --subject two-sum --workspace ./examples
relay-judge ./examples/two_sum.py
relay-judge --stress ./examples/two_sum.py
relay-judge run ./examples/two_sum.py --stress --json
relay-judge ./examples/sort_the_stack.c --cc clang
```

L'inference du sujet a partir du nom du fichier est tolerante:

- `two_sum.py`
- `two-sum.py`
- `Two Sum.py`

peuvent tous matcher le sujet attendu tant qu'il n'y a pas d'ambiguite.

## Options

### Commande `list`

Liste les sujets charges depuis le dossier `subjects`.

Options:

- `--subjects-dir <path>`: remplace le dossier de sujets detecte automatiquement

Exemple:

```bash
relay-judge list --subjects-dir ./subjects
```

### Commande `run`

Lance l'evaluation d'une soumission Python ou C.

Options:

- `--subject <id|path>`: id du sujet ou chemin vers un `subject.json`
- `--subjects-dir <path>`: dossier des sujets si `--subject` est un id
- `--workspace <path>`: dossier dans lequel chercher le fichier attendu du sujet
- `--submission <path>`: chemin explicite vers le fichier source a evaluer
- `--python <bin>`: interpreteur Python a utiliser
- `--cc <bin>`: compilateur C a utiliser
- `--json`: sortie JSON
- `--detailed`: sortie terminal plus detaillee
- `--stress`: remplace les tests normaux par la suite stress generee en memoire

Notes:

- si `--submission` est absent, le juge cherche `spec.file_name` dans `--workspace`
- si `--subject` est absent mais qu'un fichier `.py` ou `.c` est fourni en argument positionnel, le sujet est deduit du nom du fichier
- `--json`, `--detailed` et `--stress` fonctionnent aussi en trailing flags, par exemple `relay-judge run ./examples/two_sum.py --stress --json`

Exemples:

```bash
relay-judge run --subject two-sum --workspace ./examples
relay-judge run --subject two-sum --submission ./examples/two_sum.py
relay-judge run ./examples/two_sum.py --json
relay-judge run ./examples/two_sum.py --stress --detailed
relay-judge run --subject ./subjects/two-sum/subject.json --submission ./examples/two_sum.py
relay-judge run --subject two-sum --workspace ./examples --python python3.12
relay-judge run ./examples/sort_the_stack.c --cc clang
```

### Mode direct `relay-judge <file.py|file.c>`

Equivalent a un `run` avec inference du sujet par nom de fichier.

Options supportees dans ce mode:

- `--subjects-dir <path>`
- `--python <bin>`
- `--cc <bin>`
- `--json`
- `--detailed`
- `--stress`

Exemples:

```bash
relay-judge ./examples/two_sum.py
relay-judge --stress ./examples/two_sum.py
relay-judge --json --python python3 ./examples/two_sum.py
relay-judge --json --cc clang ./examples/sort_the_stack.c
```

### Mode interactif

Si le programme est lance sans argument dans un terminal interactif:

- il liste les sujets disponibles
- demande un numero de sujet
- demande un workspace
- lance l'evaluation avec le nom de fichier attendu par le sujet

## Sorties Et Codes De Retour

Statuts possibles:

- `PASSED`: tous les tests passent
- `FAILED`: la soumission repond mais au moins un test est faux
- `RUNTIME_ERROR`: la soumission leve une exception
- `TIMEOUT`: le process Python ou C depasse la limite
- `LOAD_ERROR`: le fichier ne peut pas etre charge ou le symbole attendu manque

Codes de retour:

- `0`: passed
- `1`: failed
- `2`: runtime error
- `3`: timeout
- `4`: load error
- `64`: erreur d'usage CLI
- `70`: erreur interne du juge

Formats de sortie:

- sortie terminal compacte par defaut
- `--detailed`: feuille jury plus verbeuse en mode normal
- `--json`: rapport JSON; en mode stress, le payload contient `"mode": "stress"`

## Resolution Des Chemins

Par defaut, le juge cherche `subjects/` dans plusieurs emplacements:

- a cote du binaire
- dans le repertoire courant
- dans `./relay-judge/subjects`

Tu peux toujours forcer ce comportement avec `--subjects-dir`.

## Consignes Etudiants

Pour chaque exercice, vous devez rendre un fichier source correspondant au langage du sujet.

Regles:

- le nom du fichier doit etre exactement celui attendu
- le fichier doit contenir une fonction callable avec le bon nom
- la fonction doit accepter les bons parametres
- la fonction doit retourner le resultat attendu
- le fichier doit etre du Python valide
- le programme ne doit pas planter pendant l'execution

Exemple:

- sujet `two-sum`
- fichier attendu: `two_sum.py`
- fonction attendue: `def two_sum(nums, target):`

Important:

- il faut definir la fonction, pas seulement ecrire du code en haut du fichier
- evitez les `input()` et les `print()` inutiles
- ne changez ni le nom du fichier ni le nom de la fonction
- faites une solution assez efficace pour passer les tests `perf`

Erreurs typiques:

- mauvais nom de fichier
- faute dans le nom de fonction
- erreur de syntaxe
- exception a l'execution
- mauvais format de retour

## Regeneration Des Sujets

Les `subject.json` sont generes a partir de [`scripts/generate_subjects.py`](./scripts/generate_subjects.py).

Pour regenerer les fixtures apres modification:

```bash
python3 scripts/generate_subjects.py
```

Les workspaces de variantes d'exemple sont generes par [`scripts/generate_example_variants.py`](./scripts/generate_example_variants.py).

Pour regenerer les variantes:

```bash
python3 scripts/generate_example_variants.py
```

Pour lancer un smoke test CLI rapide sur les statuts principaux:

```bash
./scripts/smoke_variants.sh
```

## Variantes De Soumissions

Le dossier `examples/variants/` contient plusieurs workspaces pour tester les differents statuts du judge:

- `slow`: solutions correctes mais volontairement lentes; a tester de preference avec `--stress`
- `wrong`: solutions qui compilent mais rendent une mauvaise reponse
- `runtime`: solutions qui levent une exception a l'execution
- `syntax`: fichiers Python invalides pour tester les erreurs de chargement
- `timeout`: solutions qui depassent volontairement la limite de temps

Exemples:

```bash
go run ./cmd/relay-judge run --subject two-sum --workspace ./examples/variants/wrong
go run ./cmd/relay-judge run --subject two-sum --workspace ./examples/variants/runtime
go run ./cmd/relay-judge run --subject two-sum --workspace ./examples/variants/syntax
go run ./cmd/relay-judge run --subject two-sum --workspace ./examples/variants/timeout
go run ./cmd/relay-judge --stress ./examples/variants/slow/two_sum.py
```
