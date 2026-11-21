# Stills mapping

Eight NPA checkpoints. One still per phase. Cover photo uses **release-point**.

| # | Phase id | Canonical file | Also matches |
|---|----------|----------------|--------------|
| 01 | `balance` | `balance.jpg` | `01-balance-posture.jpg`, names with *balance* / *posture* |
| 02 | `lift` | `lift.jpg` | `02-lift-shift.jpg`, *lift* / *shift* |
| 03 | `thrust` | `thrust.jpg` | `03-thrust.jpg`, *thrust* |
| 04 | `equal-opposite` | `equal-opposite.jpg` | `04-equal-opposite.jpg`, *equal* + *opposite* |
| 05 | `delayed` | `delayed.jpg` | `05-delayed-hips.jpg`, *delayed* / *hips* |
| 06 | `swivel` | `swivel.jpg` | `06-swivel-stabilize.jpg`, *swivel* / *stabilize* |
| 07 | `stack-track` | `stack-track.jpg` | `07-stack-track.jpg`, *stack* / *track* |
| 08 | `release-point` | `release-point.jpg` | `08-release-follow.jpg`, *release* / *follow* |

Accepted types: `.jpg` `.jpeg` `.png` `.webp`.

Prefer canonical filenames in the stills folder. If names differ, the generator scores number prefix + keywords. Override with `--map`:

```bash
--map balance=~/stills/a.jpg,lift=~/stills/b.jpg,thrust=~/stills/c.jpg,equal-opposite=~/stills/d.jpg,delayed=~/stills/e.jpg,swivel=~/stills/f.jpg,stack-track=~/stills/g.jpg,release-point=~/stills/h.jpg
```

Portrait stills (lift, equal-opposite, swivel, stack-track) must display full frame. Do not crop with `object-fit: cover`.
