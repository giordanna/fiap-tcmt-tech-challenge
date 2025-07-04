const { PubSub } = require('@google-cloud/pubsub');

async function testarPublicacao() {
  console.log('🔍 Iniciando teste de publicação no Pub/Sub...');
  // Configurar para usar o emulador
  const pubsub = new PubSub({
    projectId: 'fiap-tech-challenge-grupo-21',
    apiEndpoint: 'localhost:8085' // Endpoint do emulador Pub/Sub
  });

  const topico = 'gerar-recomendacoes';
  
  try {
    // Verificar se o tópico existe, senão criar
    const [existe] = await pubsub.topic(topico).exists();
    if (!existe) {
      await pubsub.createTopic(topico);
      console.log(`Tópico ${topico} criado.`);
    }

    // Dados de teste
    const dadosTeste = {
      timestamp: new Date().toISOString(),
      origem: 'script-teste',
      mensagem: 'Teste de publicação no worker',
      forcarProcessamento: true
    };

    // Publicar mensagem
    const dadosBuffer = Buffer.from(JSON.stringify(dadosTeste));
    const mensagemId = await pubsub.topic(topico).publish(dadosBuffer);
    
    console.log(`✅ Mensagem publicada com sucesso! ID: ${mensagemId}`);
    console.log('📄 Dados enviados:', dadosTeste);
    
  } catch (erro) {
    console.error('❌ Erro ao publicar mensagem:', erro);
  }
}

testarPublicacao();