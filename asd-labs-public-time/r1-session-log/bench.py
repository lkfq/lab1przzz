# bench.py   запуск: python bench.py        (обычный вход)
#            python bench.py worst          (худшие входы)
import subprocess, sys, re, random, statistics

EXE = "solution.exe"
RUNS = 5

def gen(lines, mode, path):
    rng = random.Random(42)
    out = []
    if mode == "normal":              # lines/2 пар, 1000 пользователей
        t = 0
        for i in range(lines // 2):
            u = f"u{i % 1000}"
            d = rng.randint(1, 100)
            out += [f"LOGIN {u} {t}", f"LOGOUT {u} {t + d}"]
            t += d + 1
    elif mode == "worst_mem":         # одни LOGIN, разные имена
        out = [f"LOGIN u{i} {i}" for i in range(lines)]
    elif mode == "worst_time":        # половина LOGIN, потом BEGIN/COMMIT
        k = lines // 2
        out = [f"LOGIN u{i} {i}" for i in range(k)]
        out += ["BEGIN", "COMMIT"] * ((lines - k) // 2)
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write("\n".join(out) + "\n")

def once(path):
    r = subprocess.run([EXE, path, "out.txt"], capture_output=True,
                       text=True, encoding="utf-8")
    m = re.search(r"time:\s*([\d.]+)\s*ms", r.stdout + r.stderr)
    if not m:
        sys.exit("Не нашёл строку 'Время работы' в выводе программы")
    return float(m.group(1))

def measure(lines, mode):
    gen(lines, mode, "in.txt")
    once("in.txt")                                  # прогрев, отбрасываем
    ts = [once("in.txt") for _ in range(RUNS)]
    return statistics.median(ts), min(ts), max(ts)

if len(sys.argv) < 2:
    print(f"{'n':>8} {'медиана':>10} {'мин':>10} {'макс':>10}")
    for pairs in (1000, 5000, 10000, 50000, 100000):
        med, lo, hi = measure(pairs * 2, "normal")
        print(f"{pairs:>8} {med:>10.2f} {lo:>10.2f} {hi:>10.2f}")
else:
    print(f"{'n':>8} {'обычный':>10} {'worst_time':>12} {'отношение':>10}")
    for n in (4000, 8000, 16000, 32000):
        a = measure(n, "normal")[0]
        b = measure(n, "worst_time")[0]
        print(f"{n:>8} {a:>10.2f} {b:>12.2f} {b / a:>10.2f}")
    print()
    print(f"{'n':>8} {'обычный':>10} {'worst_mem':>12} {'отношение':>10}")
    for n in (10000, 50000, 100000):
        a = measure(n, "normal")[0]
        b = measure(n, "worst_mem")[0]
        print(f"{n:>8} {a:>10.2f} {b:>12.2f} {b / a:>10.2f}")