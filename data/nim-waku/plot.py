import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import typer


def read_file(ifile):
    df = pd.read_table(ifile, delimiter=',', lineterminator='\n')
    df.dropna(how='any')
    return df

def case_1(value):
  if value == '--':
    return np.nan
  elif value[-3:] == 'GiB':
    return float(value[:-3])*1024*1024*1024
  elif value[-3:] == 'MiB':
    return float(value[:-3])*1024*1024
  elif value[-2:] == 'MB':
    return float(value[:-2])*1000*1000
  elif value[-2:] == 'kB':
    return float(value[:-2])*1000
  elif value[-1:] == 'B':
    return float(value[:-1])*1
  elif value[-1:] == '%':
    return float(value[:-1])
  else:
      print("ERR: ", value)
    exit()

def main(fname: str = "input_file.table"):
    df = read_file(fname)
    df.dropna()
    df['CPU%'] = df['CPU%'].map(lambda x: case_1(x))
    df['MemUsage'] = df['MemUsage'].map(lambda x: case_1(x))
    #df['MemLimit'] = df['MemLimit'].map(lambda x: case_1(x))
    df['MEM%'] = df['MEM%'].map(lambda x: case_1(x))
    df['NetI'] = df['NetI'].map(lambda x: case_1(x))
    df['NetO'] = df['NetO'].map(lambda x: case_1(x))
    df['BlockI'] = df['BlockI'].map(lambda x: case_1(x))
    df['BlockO'] = df['BlockO'].map(lambda x: case_1(x))
    df.dropna()
    #print(df)
    df.plot()
    plt.show()

if __name__ == "__main__":
    typer.run(main)
