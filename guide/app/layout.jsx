import { Layout, Navbar } from 'nextra-theme-docs'
import { Banner, Head } from 'nextra/components'
import { getPageMap } from 'nextra/page-map'
import 'nextra-theme-docs/style.css'
import '../globals.css'
import footer from '../components/footer'
import { IconBrandGithub } from '@tabler/icons-react';

export const metadata = {
  title: 'Redigo',
  description: 'The purpose of this series is to guide you in building an in-memory database like Redis. We will be using Go as our programming language and the series will be divided into multiple parts.',
}
 
const banner = <Banner storageKey="some-key">Nextra 4.0 is released 🎉</Banner>
const navbar = (
  <Navbar
    logo={
      <div className='flex w-full items-center justify-center gap-2'>
          <p className='font-bold text-xl text-red-500 x:text-black x:dark:text-white font-mono zsft-443'>
            Redigo
          </p>
      </div>
    }
    projectLink="https://github.com/inannan423/redigo"
    projectIcon = <IconBrandGithub stroke={2} />
  />
)

 
export default async function RootLayout({ children }) {
  return (
    <html
      lang="en"
      dir="ltr"
      suppressHydrationWarning
    >
      <Head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <link rel="icon" href="/favicon.ico" />
        {/* favicon */}
        <link rel="icon" type="image/svg" href="/redigo.svg" />
      </Head>
      <body suppressHydrationWarning>
        <Layout
          // banner={banner}
          navbar={navbar}
          pageMap={await getPageMap()}
          docsRepositoryBase="https://github.com/inannan423/redigo/tree/main/guide"
          footer={footer}
        >
          {children}
        </Layout>
      </body>
    </html>
  )
}