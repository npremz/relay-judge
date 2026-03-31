# Jury Evaluation Plan

Objectif: aligner le juge automatique avec la grille jury sans pretendre automatiser les criteres qui doivent rester humains.

## Ce que le juge doit objectiver

- `Correction /40`: via les tests du groupe `core`
- `Edge cases /20`: via les tests du groupe `edge`
- `Qualite algorithmique / complexite /20`: via les groupes `perf` et `anti-hardcode`

## Ce que le juge ne doit pas trancher seul

- `Lisibilite /10`: reste manuel
- `Rapidité /10`: reste manuel, selon le rang reel d'arrivee

## Groupes de tests

- `core`: cas standards, verifies que la solution resout bien le coeur du sujet
- `edge`: cas limites et robustesse
- `anti-hardcode`: variantes qui evitent les solutions collees aux exemples
- `perf`: cas qui mettent en evidence une approche trop naive

## Suggestions de score

Le juge doit fournir des suggestions et non un verdict definitif.

### Correction /40

Basee sur le ratio de tests `core` passes:

- `0%` -> `0`
- `1%-39%` -> `10`
- `40%-69%` -> `20`
- `70%-99%` -> `30`
- `100%` -> `40`

### Edge cases /20

Basee sur le ratio de tests `edge` passes:

- `0%` -> `0`
- `1%-34%` -> `5`
- `35%-67%` -> `10`
- `68%-99%` -> `15`
- `100%` -> `20`

### Complexite /20

Basee sur `perf` et `anti-hardcode`.

- si timeout ou crash avant les tests: suggestion faible
- si les tests `perf` et `anti-hardcode` passent: suggestion forte
- le jury garde la validation finale apres lecture du code

Barème suggere:

- signal nul -> `0`
- signal faible -> `5`
- signal moyen -> `10`
- signal bon -> `15`
- signal excellent -> `20`

## Rapport jury attendu

Le binaire doit afficher:

- le statut global
- le detail par groupe
- une suggestion `Correction /40`
- une suggestion `Edge cases /20`
- une suggestion `Complexite /20`
- un rappel explicite que `Lisibilite` et `Rapidité` restent manuels
