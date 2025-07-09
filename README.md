# Tech Challenge POC: Módulo de recomendações

## Instruções das Functions:

1. Instale as dependências:
```bash
cd functions
npm install
```

2. Compile o código TypeScript:
```bash
npm run build
```

3. Para desenvolvimento com reload automático:
```bash
npm run build:watch
```

4. Para servir localmente:
```bash
npm run serve
```

5. Para realizar o deploy manualmente:
```bash
npm run deploy
```

## Instruções da geração do datalake

1. Crie um amviente virtual:
```bash
cd scripts
python -m venv venv
```

2. Ative o ambiente virtual (opcional):
```bash
# windows
.\venv\Scripts\activate

# linux/mac
source venv/bin/activate
```

3. Instale as dependências:
```bash
pip install -r requirements.txt
```

4. Execute o script
```bash
python gerar_datalake.py
```