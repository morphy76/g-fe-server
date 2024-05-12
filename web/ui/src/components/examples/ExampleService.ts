import { useMutation, useQuery, useQueryClient } from 'react-query';
import { ps_client } from '../../features/axios';
import { loggerFor } from '../../features/react-logger';

export type Example = {
  name: string;
  age: number;
};

export const listExamples = async (): Promise<Example[]> => {
  loggerFor('ExampleService').debug('listExamples');
  const res = await ps_client.get('/example');
  const sortedData = res.data.sort((a: Example, b: Example) => a.name.localeCompare(b.name));
  return sortedData;
};

export const useListExampleQuery = () => {
  return useQuery<ReadonlyArray<Example>, Error>("examples", listExamples);
};

export const getExample = async (name: string | null): Promise<Example | null> => {
  if(!name) return null;
  loggerFor('ExampleService').debug('getExample', name);
  const res = await ps_client.get(`/example/${name}`);
  return res.data;
};

export const useGetExample = (selected: string | null) => {
  return useQuery<Example | null, Error>({
    queryKey: selected ?? 'example',
    queryFn: () => getExample(selected),
    enabled: !!selected
  });
};

export const createExample = async (example: Example): Promise<void> => {
  loggerFor('ExampleService').debug('createExample', example);
  await ps_client.post('/example', example);
};

export const useCreateExample = () => {
  const queryClient = useQueryClient();
  return useMutation((example: Example) => createExample(example), {
    mutationKey: 'createExample',
    onSuccess: () => {
      queryClient.invalidateQueries('examples');
    }
  });
};

export const replaceExample = async (name: string, example: Example): Promise<void> => {
  loggerFor('ExampleService').debug('replaceExample', name, example);
  await ps_client.put(`/example/${name}`, example);
};

export const useReplaceExample = (name: string) => {
  const queryClient = useQueryClient();
  return useMutation((example: Example) => replaceExample(name, example), {
    mutationKey: 'replaceExample',
    onSuccess: () => {
      queryClient.invalidateQueries('examples');
    },
  });
};

export const deleteExample = async (name: string): Promise<void> => {
  loggerFor('ExampleService').debug('deleteExample', name);
  await ps_client.delete(`/example/${name}`);
};

export const useDeleteExample = () => {
  const queryClient = useQueryClient();
  return useMutation((name: string) => deleteExample(name), {
    mutationKey: 'deleteExample',
    onSuccess: () => {
      queryClient.invalidateQueries('examples');
    },
  });
};
