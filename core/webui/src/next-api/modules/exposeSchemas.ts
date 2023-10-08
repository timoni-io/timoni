import { ApiShema, createModule } from "../modules";

export type ExposeSchemas<T extends ApiShema> = {
  [K in keyof T]: T[K];
};

export const ExposeSchemasModule = createModule({
  addProperty(obj, schema) {
    return [obj, schema];
  },
});
