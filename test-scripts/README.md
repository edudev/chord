# Test Scripts

## Installation
```bash
./setup.sh
```

## Usage
```bash
module load slurm

sbatch -N<node count x 2> ./test-sbatch.sh
```

## Watching output
```bash
watch -n1 cat slurm-<job id>.out
```

For the test results, you can check the files stated in the output.
