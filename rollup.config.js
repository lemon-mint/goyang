import typescript from '@rollup/plugin-typescript';
import { minify } from 'rollup-plugin-swc-minify';

export default {
    input: 'src/index.ts',
    output: {
        dir: 'dist',
        format: 'esm',
    },
    plugins: [
        typescript(),
        minify(),
    ]
};
