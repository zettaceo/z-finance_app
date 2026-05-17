Contexto — Pre-cadastro institucional

Objetivo
- Implementar sistema completo de pre-cadastro institucional com dupla verificacao (email e telefone).

Escopo
- Criar entidade/tabela pre_registrations e logs de verificacao.
- Endpoints para iniciar pre-cadastro, verificar email, verificar telefone e consultar status.
- Antifraude: limites de tentativas, expiracao e bloqueio temporario.
- Preparar fluxo futuro de conversao sem criar conta agora.

Decisoes
- Tokens e codigos armazenados com hash.
- Expiracao padrao: 48h (configuravel).
- Tentativas maximas e cooldown configuraveis.

Pendencias
- Mapear camadas atuais (entity, repo, usecase, handler, migrations). (feito)
- Implementar schema e migracao. (feito)
- Implementar repositorios e usecases. (feito)
- Implementar handlers e rotas. (feito)
- Atualizar docs e logs de continuidade. (pendente)
- Executar testes e registrar resultados. (pendente)
