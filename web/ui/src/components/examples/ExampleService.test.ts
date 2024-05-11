import { ps_client } from "../../features/axios";
import MockAdapter from "axios-mock-adapter";
import { listExamples } from "./ExampleService";

describe('Example resource', () => {

  let mock: MockAdapter | undefined;

  beforeAll(() => {
    mock = new MockAdapter(ps_client);
  });

  afterEach(() => {
    mock?.reset();
  });

  it('should list examples', async () => {

    mock?.onGet('/example').reply(200, [
      { name: "John", age: 30 },
      { name: "Jane", age: 25 }
    ]);

    const examples = await listExamples();

    expect(examples).toBeDefined();
    expect(Array.isArray(examples)).toBeTruthy();
    expect(examples.length).toBe(2);
  });
});
