# Benchmark results

5-layer networks, with same # of units per layer: SMALL = 25; MEDIUM = 100; LARGE = 625; HUGE = 1024; GINORM = 2048, doing full learning, with default params, including momentum, dwtnorm, and weight balance.

Results are total time for 1, 2, 4 threads, on my macbook.

## C++ Emergent

```
* SMALL:   2.383   2.248    2.042
* MEDIUM:  2.535   1.895    1.263
* LARGE:  19.627   8.559    8.105
* HUGE:   24.119  11.803   11.897
* GINOR:  35.334  24.768   17.794
```

## Go emergent 6/2019 after a few bugfixes, etc: significantly faster!

```
* SMALL:   1.46   3.63   3.96
* MEDIUM:  1.87   2.46   2.32
* LARGE:  11.38   8.48   6.03
* HUGE:   19.53   14.52   11.29
* GINOR:  26.93   20.66   15.71
```

now really just as fast overall, if not faster, than C++!

note: only tiny changes after adding IsOff check for all neuron-level computation.

## Go emergent, per-layer threads, thread pool, optimized range synapse code

```
* SMALL:   1.486   4.297   4.644
* MEDIUM:  2.864   3.312   3.037
* LARGE:  25.09   20.06   16.88
* HUGE:   31.39   23.85   19.53
* GINOR:  42.18   31.29   26.06
```

also: not too much diff for wt bal off!

## Go emergent, per-layer threads, thread pool

```
* SMALL:  1.2180    4.25328  4.66991
* MEDIUM: 3.392145  3.631261  3.38302
* LARGE:  31.27893  20.91189 17.828935
* HUGE:   42.0238   22.64010  18.838019
* GINOR:  65.67025  35.54374  27.56567
```

## Go emergent, per-layer threads, no thread pool (de-novo threads)

```
* SMALL:  1.2180    3.548349  4.08908
* MEDIUM: 3.392145  3.46302   3.187831
* LARGE:  31.27893  22.20344  18.797924
* HUGE:   42.0238   29.00472  24.53498
* GINOR:  65.67025  45.09239  36.13568
```

# Per Function 

Focusing on the LARGE case:

C++: `emergent -nogui -ni -p leabra_bench.proj epochs=5 pats=20 units=625 n_threads=1`

```
BenchNet_5lay timing report:
function  	time     percent 
Net_Input     8.91    43.1
Net_InInteg	   0.71     3.43
Activation    1.95     9.43
Weight_Change 4.3     20.8
Weight_Updt	   2.85    13.8
Net_InStats	   0.177    0.855
Inhibition    0.00332  0.016
Act_post      1.63     7.87
Cycle_Stats	   0.162    0.781
    total:	   20.7
```

Go: `./bench -epochs 5 -pats 20 -units 625 -threads=1`

```
TimerReport: BenchNet, NThreads: 1
    Function Name  Total Secs    Pct
    ActFmG         2.121      8.223
    AvgMaxAct      0.1003     0.389
    AvgMaxGe       0.1012     0.3922
    DWt            5.069     19.65
    GeFmGeInc      0.3249     1.259
    InhibFmGeAct   0.08498    0.3295
    QuarterFinal   0.003773   0.01463
    SendGeDelta   14.36      55.67
    WtBalFmWt      0.1279     0.4957
    WtFmDWt        3.501     13.58
    Total         25.79
```

```
TimerReport: BenchNet, NThreads: 1
    Function Name    Total Secs    Pct
    ActFmG           2.119     7.074
    AvgMaxAct        0.1        0.3339
    AvgMaxGe        0.102     0.3407
    DWt             5.345     17.84
    GeFmGeInc        0.3348     1.118
    InhibFmGeAct     0.0842     0.2811
    QuarterFinal     0.004    0.01351
    SendGeDelta     17.93     59.87
    WtBalFmWt        0.1701     0.568
    WtFmDWt        3.763     12.56
    Total         29.96
```

* trimmed 4+ sec from SendGeDelta by avoiding range checks using sub-slices
* was very sensitive to size of Synapse struct


