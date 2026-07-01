import { useState } from 'react';
import {
  Dialog, DialogTrigger, ActionButton, Form, TextField, TextArea,
  Button, Content, Heading, Divider, Flex,
} from '@adobe/react-spectrum';
import Add from '@spectrum-icons/workflow/Add';

interface Props {
  onCreate: (data: { name: string; business_desc?: string; seed_keywords?: string[] }) => Promise<any>;
}

export default function ProjectForm({ onCreate }: Props) {
  const [name, setName] = useState('');
  const [desc, setDesc] = useState('');
  const [keywordStr, setKeywordStr] = useState('');
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setLoading(true);
    const keywords = keywordStr.split('\n').map(s => s.trim()).filter(Boolean);
    try {
      await onCreate({ name, business_desc: desc, seed_keywords: keywords });
      setOpen(false);
      setName('');
      setDesc('');
      setKeywordStr('');
    } finally {
      setLoading(false);
    }
  };

  return (
    <DialogTrigger isOpen={open} onOpenChange={setOpen}>
      <ActionButton><Add /></ActionButton>
      <Dialog>
        <Heading>Novo Projeto</Heading>
        <Divider />
        <Content>
          <Form>
            <TextField label="Nome" value={name} onChange={setName} isRequired autoFocus />
            <TextArea label="Descrição do negócio" value={desc} onChange={setDesc} />
            <TextArea
              label="Seed Keywords (uma por linha)"
              value={keywordStr}
              onChange={setKeywordStr}
              height="size-2000"
            />
          </Form>
        </Content>
        <Flex gap="size-100" justifyContent="end">
          <Button variant="secondary" onPress={() => setOpen(false)}>Cancelar</Button>
          <Button variant="accent" onPress={submit} isDisabled={!name || loading}>
            {loading ? 'Criando...' : 'Criar'}
          </Button>
        </Flex>
      </Dialog>
    </DialogTrigger>
  );
}
